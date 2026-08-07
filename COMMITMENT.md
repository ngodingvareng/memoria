# Commitment — User-Perspective Walkthrough

> **Status: long-horizon. Not scheduled, not being built.**
>
> Commitment is the one feature in Memoria that needs a long design
> conversation before a line of it is written. It is the only mode that
> judges the user, it touches the scheduler, the notification system, the
> archive surfaces, and the privacy model at once, and several of its
> rules are still undecided (see [Open questions](#open-questions)).
>
> Nothing here is a commitment to build. This document exists so the
> feature can be *reasoned about* without being started, and so the work
> that eventually happens starts from settled questions instead of
> discovering them halfway through. Treat everything below as design
> notes, not a backlog.

[`FEATURES.md`](FEATURES.md#commitment) defines what Commitment **is** —
its opt-in contract, its strictness levels, its firewall. This document
covers what it **feels like to use**, walked through scenario by
scenario, because most of the unresolved design problems only become
visible once you follow a real user through a real week.

---

## The core difference: who creates the empty page

One sentence explains the whole feature.

- **Manual** — the user creates a Moment. Nothing was waiting. There is
  no state.
- **Commitment** — *the app* creates an empty Moment on a schedule, and
  then waits for the user to fill it in. That Moment has a state: **due**
  → **kept** / **missed**.

So Commitment is not a reminder. A reminder tells you something. A
Commitment has **already created an entry that sits there open**, and
records whether you came to fill it. That is where the psychological
weight comes from, and it is exactly what the user asked for when they
opted in.

---

## Setup

A Thread named "Morning Run" already exists and holds five manually
captured Moments. The user opens Thread settings → **Add commitment**,
reads the deliberately blunt consent screen (full text in
`FEATURES.md`), and sets three things:

| Setting | Example |
|---|---|
| Recurrence | every day at 06:00 |
| Confirmation window | 6 hours (until 12:00) |
| Strictness | Gentle / Standard / Strict |

After saving, the Thread does not visibly change. Existing Moments stay
ordinary Moments. The only difference is that starting tomorrow, entries
appear on their own.

---

## Scenario A — The normal path (kept)

- **06:00** — the system creates a Moment: `status=due`, `due_at=06:00`,
  `occurred_at=06:00`, no content. A *Moment due* notification goes out:
  "How did 'Morning Run' go?"
- **07:10** — the user opens the app. In the Thread there is one card
  unlike the others: a card that is waiting. They add a photo and the
  note "5k, new route."
- Saving sets `status=kept`, `captured_at=07:10`.
- They nudge `occurred_at` to 06:15 because the run actually started
  late. Allowed — `occurred_at` always belongs to the user.

Once saved, this Moment is **indistinguishable** from a manual one in
Album, Echoes, or Search. Its status lives only in the adherence
surfaces.

---

## Scenario B — The window closes (missed)

The user forgets. **12:00** passes and the timeout worker flips **due** →
**missed**.

What the user experiences:

- With the *Moment missed* notification **on** (off by default): one flat
  notification — "'Morning Run' wasn't recorded today." No "again," no
  adverb doing moral work.
- With it off: complete silence that day. They find out in the monthly
  adherence review — which is the design, because "learn from your past"
  requires the mistake to actually be in the past.

This has an architectural consequence worth stating plainly: because the
Commitment Firewall forbids **missed** Moments from Album, Echoes,
Search, and a Thread's browsable timeline, **the commitment record needs
its own room.** It cannot be "the Thread timeline with a filter." In
practice the user reaches adherence through a separate surface — an
Adherence tab on the Thread, plus aggregates in Rhythms and Recap.

---

## Scenario C — Recorded late, and what strictness actually controls

The user opens the app at 20:00 and fills in the Moment that already
went **missed**.

| Strictness | Result |
|---|---|
| **Gentle** | Counts as kept. No marker. |
| **Standard** | Counts as kept, noted as late in the adherence review. |
| **Strict** | Does not count. But the content is **saved in full**. |

The thing most easily got wrong: **strictness affects the scorecard, not
the archive.** A photo recorded at 20:00 under Strict still appears in
next year's Album like any other Moment. Only the number reflects the
miss.

This scenario is what exposes [C2](#c2) below.

---

## Scenario D — Two Commitments on one Thread

A "Fitness" Thread carries Morning Run (06:00, Strict) and Evening Walk
(19:00, Gentle).

Each day produces two independent **due** Moments — own window, own
strictness, own adherence rate. In the Thread timeline they look like two
Moments on that day. Only the adherence review separates them: "Morning
Run 18/30 · Evening Walk 27/30."

The schema already supports this (`moments.commitment_id`, plus a
composite FK guaranteeing the linked Commitment belongs to the same
Thread). What it does not yet support: `confirmation_timeout_minutes` and
strictness still live on `threads`, so these two Commitments **cannot
currently differ in how demanding they are** — tracked in
`api/TODO.md` Tier 5.

---

## Scenario E — Changing the schedule mid-flight

On August 10 the user changes 06:00 → 07:00.

- Moments for August 1–9 keep `due_at` 06:00. Nothing is rewritten.
- From August 10 onward, 07:00 applies.
- July's adherence is still computed against the rule **actually in force
  at the time** — this is what `commitment_histories` is for.

The sentence that explains it to a user: *changing the alarm doesn't
change what happened.*

---

## Scenario F — Pause (a holiday)

The user pauses for two weeks. No Moments are generated.

The psychologically critical part: **those days leave the denominator;
they do not enter as misses.** The adherence review shows it as a neutral
fact — "paused Aug 10–24" — not a hole in the graph. If pausing felt like
damaging the record, users would not pause; they would let the Commitment
rot and then avoid the app, and Principle 1 loses.

---

## Scenario G — The exit ramp

Three kept out of the last twenty occurrences. The app offers the way out
before the user has to find one:

> "Morning Run: 3 of the last 20. Want to change the schedule, loosen it,
> or pause it for a while?"

It appears when the user opens the Thread, or in the adherence review —
**not** as an evening notification, not as a red banner. All four options
(change schedule / loosen / pause / turn off) are presented as equals.
"Turn off" must not read as the give-up option tucked in a corner.

This is the most important safeguard in the feature: it turns the moment
of maximum shame into the moment the app is most obviously on the user's
side.

---

## Scenario H — Turning Commitment off entirely

The Thread reverts to an ordinary manual Thread. Everything already
recorded stays exactly as it is.

The history itself should stay viewable as a closed chapter — *"January
– August 2025: you committed to this, and kept 210 of 240."* That is
precisely the nostalgia value the mode was built for, and it survives the
Commitment ending.

What happens to still-**due** Moments at the moment of shutdown is
undecided — see [C3](#c3).

---

## Scenario I — Travel

A Commitment set to "06:00, Asia/Jakarta." The user flies to Tokyo.

Because the timezone is stored on the Commitment, the Moment comes due at
06:00 Jakarta — 08:00 local time in Tokyo. Whether that is right depends
on what the user meant, which is [C4](#c4).

Separately and independently: `occurred_at` still stores the local UTC
offset where it happened, so a run in Tokyo on August 4 stays on August 4
forever. That part is settled (`FEATURES.md` § Timezones and travel).

---

## Scenario J — Looking back a year later

This is why the mode exists, and where the Firewall is most visible:

- **Album, March 2025** → the mornings that happened. Not a column of
  empty days.
- **Echoes** → surfaces a real run. Never "a year ago you skipped."
- **Rhythms** → "2025: 210 runs. Weekdays 85%, weekends 20%."
- **Recap, March** → consistency, in aggregate.

That second line in Rhythms is the valuable one. **Most missed
commitments are schedule design errors, not character failures.** Someone
who reliably misses weekends does not need a 62% score; they need to know
their weekend schedule was never realistic.

---

## Open questions

These block implementation. Each one is a product call, not a technical
detail, and each was found by walking the scenarios above rather than by
reading the spec.

### C1 — Does enabling a Commitment generate a Moment immediately?

Turning a Commitment on at 20:00 for a 06:00 daily schedule: does today
get an occurrence, or does generation start at the next one? Almost
certainly the next one — handing a new user a **due** Moment that is
already impossible to answer is the worst possible first impression — but
it is unstated.

### C2 — A missed Moment that has content

Under Strict, a late-recorded Moment stays **missed** *and* holds a real
photo and note. The Firewall says **missed** Moments never appear in
archive surfaces. Applied literally, that deletes a genuine memory from
Album because the user was slow.

`status` is currently answering two unrelated questions: *"did this
happen?"* and *"did this count as adherence?"* They have to be separated.
This is the concrete case that makes `FEATURES.md` Open Decision 5
(moving Commitment outcomes into their own table) a necessity rather than
a refinement.

### C3 — Still-due Moments when a Commitment is paused or turned off

If they become **missed**, the app punishes the user for stopping, which
poisons the exit ramp — the single safeguard that matters most. A third
outcome (something like *cancelled*) that never enters the denominator is
needed.

### C4 — Which timezone a Commitment fires in

Fixed to the Commitment's timezone is correct for *"my morning run at
home."* Following the device is correct for *"my morning run wherever I
am."* Both are reasonable schedules. This may have to be a per-Commitment
choice rather than a global rule.

### C5 — Manual capture on a Thread with an active Commitment

**The largest one.** The user finishes their run at 05:30, before the
06:00 Moment is generated, and taps the ordinary Capture button. At 06:00
the system creates a **due** Moment anyway; at 12:00 it marks it
**missed** — while the user demonstrably ran.

The schema permits both to coexist (`commitment_id` is nullable), but the
product has not decided whether a manual capture may *claim* an open
**due** Moment. This must be settled before the scheduler is written: it
is a bug every user would hit in their first week.

### C6 — A confirmation window longer than the interval

Every 6 hours with an 8-hour window produces overlapping **due** Moments.
Confusing in the UI ("which one am I answering?"), and
`uq_moments_thread_id_due_at` does not prevent it because the `due_at`
values genuinely differ. The cheap fix is validating `window ≤ interval`
at Commitment creation.

### C7 — Backdating `occurred_at` on a Commitment Moment

The user fills in Wednesday's Moment but moves `occurred_at` to Monday.
Does that affect Monday's already-**missed** Moment? The sane answer is
no — `due_at` and `occurred_at` are different axes, and letting users
move `occurred_at` to repair their record would make the record
meaningless.

### Already tracked elsewhere

- **`FEATURES.md` Open Decision 2a** — does loosening strictness apply
  retroactively?
- **`api/TODO.md` Tier 5** — `moments.status` is `NOT NULL` (manual
  Moments are forced to carry a state), `confirmation_timeout_minutes`
  lives on `threads` instead of per-Commitment, and
  `GetSettlingTimeStats` measures Commitment punctuality rather than Time
  to Tell.
- **`api/TODO.md` Tier 5** — whether changing the confirmation window
  applies retroactively to old **due** Moments.
