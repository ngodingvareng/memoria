# Schema Review — Potential Issue Scenarios

This document collects the risks and edge cases identified while reviewing the
database schema (`schema.sql`) for the activity-tracking app. Some have
already been addressed at the schema level; others are business-logic /
application-layer concerns that the schema alone cannot enforce, and are
listed here so they aren't lost.

## Summary

| # | Scenario | Area | Status |
|---|----------|------|--------|
| 1 | Confirmation vs. timeout job race | Concurrency | Mitigated at query layer |
| 2 | Duplicate item generation from multiple workers | Concurrency | Mitigated at query layer |
| 3 | Concurrent schedule edits | Concurrency | Open |
| 4 | Schedule linked to a different activity | Data integrity | **Resolved** (composite FK) |
| 5 | `is_fixed_schedule` out of sync with `activity_schedules` | Data integrity | Open |
| 6 | Editing a schedule via delete+insert breaks history links | Data integrity | Open (process discipline) |
| 7 | Retroactive effect of changing `confirmation_timeout_minutes` | Lifecycle | Open |
| 8 | Soft-deleted user's sessions stay valid | Lifecycle / security | Open |
| 9 | Orphaned image files after hard delete | Lifecycle | Open |
| 10 | Verification token/OTP replay | Security | Open |
| 11 | Unrestricted `activity_items.status` transitions | Data integrity | Open (product decision) |
| 12 | No user timezone for date bucketing | Missing field | **Resolved** (`users.timezone`) |

---

## 1. Confirmation vs. Timeout Job Race

**Scenario:** An item is `awaiting`. The timeout worker evaluates it as
overdue at nearly the same moment the user taps "done" in the app. If the
worker blindly runs `UPDATE ... SET status = 'missed' WHERE id = ?`, it could
overwrite a confirmation that just landed a moment earlier (or vice versa).

**Mitigation in place:** `ConfirmActivityItem` and `MarkOverdueItemsAsMissed`
both guard with `WHERE status = 'awaiting'`, so whichever write lands first
wins and the other becomes a no-op instead of clobbering data. This only
works as long as every write path to `status` keeps using that guard —
worth calling out in code review if a new mutation is added later.

## 2. Duplicate Item Generation From Multiple Workers

**Scenario:** If the scheduler ever runs on more than one instance (horizontal
scaling), two workers could try to generate the same occurrence for the same
activity at the same time.

**Mitigation in place:** `uq_activity_items_activity_scheduled_at` (a partial
unique index on `(activity_id, scheduled_at) WHERE deleted_at IS NULL`)
rejects the second insert. `CreateScheduledActivityItem` already uses
`ON CONFLICT ... DO NOTHING` to make that the expected, non-error outcome.

## 3. Concurrent Schedule Edits

**Scenario:** Two requests to edit the same schedule's `cron_expression`
(e.g. a double-submitted form) could race, each archiving a slightly
different "old" snapshot into `activity_schedule_histories`, producing
overlapping or inconsistent `active_from`/`active_until` windows.

**Status:** Not enforced by the schema. If this matters, consider a
`SELECT ... FOR UPDATE` on the `activity_schedules` row (or an optimistic
`updated_at`/version check) before archiving + updating.

## 4. Schedule Linked to a Different Activity — Resolved

**Scenario:** `activity_items.schedule_id` originally only referenced
`activity_schedules(id)`, with nothing checking that the schedule's
`activity_id` matched the item's own `activity_id`. A bug could link an item
to a schedule belonging to an unrelated activity.

**Resolution:** `activity_schedules` now has
`UNIQUE (activity_id, id)`, and `activity_items` enforces a composite FK:

```sql
CONSTRAINT fk_activity_items_schedule_activity
    FOREIGN KEY (activity_id, schedule_id)
    REFERENCES activity_schedules (activity_id, id)
    ON DELETE SET NULL (schedule_id)
```

Note this requires **PostgreSQL 15+** — the column-specific `SET NULL`
action is what lets only `schedule_id` be nulled out when a schedule is
deleted, instead of also nulling `activity_id` (which is `NOT NULL` and
would otherwise error).

## 5. `is_fixed_schedule` Out of Sync With `activity_schedules`

**Scenario:**
- An activity has `is_fixed_schedule = true` but zero rows in
  `activity_schedules` — auto-generation silently stops working and nothing
  surfaces that fact.
- An activity is toggled to `is_fixed_schedule = false` but old schedule
  rows are never cleaned up, and a buggy generator job that doesn't check
  the flag keeps producing items anyway.

**Status:** Not enforced by the schema — purely an application invariant.
`SetActivityFixedSchedule`'s query comment flags this, but the app must
explicitly retire/archive schedules whenever the flag flips to `false`, and
the generator job must always join on `activities.is_fixed_schedule = true`
(see `ListActiveSchedulesForGeneration`) rather than iterating
`activity_schedules` alone.

## 6. Editing a Schedule via Delete+Insert Breaks History Links

**Scenario:** If "editing a schedule" is implemented as delete-old-row +
insert-new-row (instead of `UPDATE` in place), every `activity_item` that
referenced the old row via `schedule_id` loses that link (`schedule_id`
becomes `NULL` per the `ON DELETE SET NULL` rule), even though
`activity_schedule_histories` still records the cron change.

**Recommendation:** Always edit via `UPDATE` on the same row
(`UpdateActivitySchedule` in the query set does this), archiving the old
values into `activity_schedule_histories` *before* the update, in the same
transaction.

## 7. Retroactive Effect of Changing `confirmation_timeout_minutes`

**Scenario:** The timeout job computes each item's deadline live as
`scheduled_at + activities.confirmation_timeout_minutes` rather than storing
a fixed deadline per item. If a user shortens the timeout (e.g. from 1440 to
60 minutes), every still-`awaiting` item — including old ones created under
the previous, more lenient setting — could flip to `missed` on the very next
job run.

**Status:** Open product decision. If only new items should be affected,
consider storing a computed `expires_at` directly on each `activity_item` at
creation time, instead of deriving it from the activity's current setting.

## 8. Soft-Deleted User's Sessions Stay Valid

**Scenario:** `users.deleted_at` gives a recovery grace period, but setting
it doesn't automatically invalidate that user's existing rows in
`user_sessions`. Anyone still holding a valid session (or a stolen one)
could keep using the app during the grace period even though the account
was "deleted."

**Recommendation:** Whenever `SoftDeleteUser` runs, also call
`DeleteAllSessionsByUserID` in the same transaction (the query file already
notes this).

## 9. Orphaned Image Files After Hard Delete

**Scenario:** `activity_images.image_path` / `activity_item_images.image_path`
are just string references to files in external storage. When a row is
hard-deleted (via cascade), Postgres has no way to delete the corresponding
file — it will accumulate as storage cost unless something else cleans it up.

**Recommendation:** Handle file deletion explicitly in application code
around the delete transaction (or run a periodic reconciliation job that
diffs storage against `image_path` values still referenced in the DB).

## 10. Verification Token/OTP Replay

**Scenario:** `user_verifications` has no "consumed" marker. If the app's
verify flow only deletes the row *after* using it, two near-simultaneous
requests using the same code could both pass validation before either
delete completes.

**Recommendation:** Either use `SELECT ... FOR UPDATE` when validating +
consuming a verification in the same transaction, or add a `consumed_at`
column and check `consumed_at IS NULL` as part of the validation query.

## 11. Unrestricted `activity_items.status` Transitions

**Scenario:** Nothing stops a `missed` item from later being flipped back to
`captured` (a late confirmation), or a `captured` item reverting to
`awaiting`. Whether late confirmation should be allowed at all is a product
decision this schema doesn't make for you.

**Recommendation:** If certain transitions should never happen, enforce them
either with a `CHECK` constraint referencing `OLD`/`NEW` via a trigger, or —
more simply — with an explicit rule in the application's update logic.

## 12. No User Timezone for Date Bucketing — Resolved

**Scenario:** The heatmap and other statistics group items by "day," but a
day boundary depends on whose timezone you're using. `occurred_at` and
`scheduled_at` are absolute instants (`TIMESTAMPTZ`); without a stored
per-user timezone, there was no reliable way to convert those instants into
"which calendar day" for that specific user.

**Resolution:** Added `users.timezone VARCHAR(50) NOT NULL DEFAULT 'UTC'`.
`GetHeatmapData` now buckets using
`(activity_items.occurred_at AT TIME ZONE users.timezone)::date`.
