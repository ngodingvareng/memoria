# Memoria Features

> Capture moments and history today to weave the memories of tomorrow.

## Philosophy

Memoria is a private memory keeper, not an activity tracker. Its value is deferred: almost nothing it does is worth much on day one, and all of it compounds by year three. Every design decision follows from that.

The archive is the product. Features are judged by whether they make the archive longer, richer, and safer to revisit — not by how much they get used this week.

### Principles

These are non-negotiable. A feature that violates one is wrong, however well it performs.

1. **The archive is the product.** Anything that raises the chance a user abandons Memoria is a bug, no matter how motivating it is in the short term. A three-year archive beats a well-optimized three-week one.
2. **Memoria never makes you feel you owe it something — unless you asked it to.** By default the app has no claim on the user: it does not scold, score, or keep count of failures. The single exception is **Commitment** mode (see Activity Commitment), which a user opts into deliberately, is told plainly what it will do, and can leave at any time. Accountability a person chose for themselves motivates; accountability imposed on them corrodes. The consent is the whole difference, which is why it has to be real rather than buried in a settings toggle.
3. **The mundane is the point.** What people miss after ten years is the commute, the old apartment, the routine coffee — not the highlights. Memoria is built to catch the ordinary, which is exactly what highlight-driven apps throw away.
4. **Nothing you recorded is ever removed by someone else's action.** Another user blocking you, leaving a circle, or deleting their account never subtracts from your archive.
5. **Sharing is relational, never broadcast.** A moment reaches someone because they were part of it, or because they share a circle with you — never because it was published to an audience.
6. **Resurfacing is a gift, and gifts can be declined.** Anything the app surfaces unprompted must be refusable, at any granularity, permanently.
7. **Capture must be cheap.** The richer the capture form, the less often people fill it in. One tap is always enough; everything else is optional and can be added later — including years later.

### Non-Goals

Memoria deliberately does not have, and will not add:

- Public profiles, follower counts, or any audience metric.
- Visible reaction/comment counts, or leaderboards of any kind.
- An algorithmic or ranked feed.
- Streaks, or any "current streak: 0" style counter. Commitment mode reports **rates** instead, and the difference is not cosmetic: a rate absorbs a single bad day, while a streak is destroyed by one — which is precisely the mechanic that turns one lapse into quitting altogether.
- Read receipts or view counts. In circles this small, being seen and not answered is itself a message — Memoria refuses to generate that message.
- Real-time engagement notifications designed to pull the user back in.
- Ads, or any use of captures for training, profiling, or targeting.

---

## Activity

The main entity that defines an activity, hobby, or routine. It acts as a container for its Captures (see below). An Activity can be personal (owned by a single user) or collaborative, owned by a Circle (see the Circle section). By default, its Captures are added manually whenever the user wants; it can optionally opt into a fixed, recurring schedule instead — see Activity Schedule below.

An Activity is a *thread of life*, not a task. "Work," "Coffee Exploring," and "Morning Run" are recurring textures of a life, and the accumulated ordinariness of a thread is what makes it worth reading a decade later. This is Memoria's most distinctive idea and the app should be presented around it.

> My Activities:
>
> - "Morning Run"
> - "Coding Together"
> - "Coffee Exploring"

## Activity Capture

A single occurrence of an Activity. This is the core of Memoria. A Capture comes into existence one of two ways:

- **Manually** — the user adds it themselves, whenever they want. This is Memoria's primary, simplest, and most important path, and the one that most directly reflects the app's philosophy: log life as it happens, on your own terms, with nothing scheduled or owed.
- **From a Commitment** — for Activities the user has deliberately put under a fixed schedule (see Activity Commitment), the system generates the Capture the moment it comes due. This is an advanced, opt-in mode layered on top of the core experience, with stricter mechanics and a deliberately different emotional character. Manual Captures can still be added freely alongside a Commitment at any time.

**A manual Capture has no state.** It is created and it exists; that is the entire model. There is nothing to confirm, nothing pending, and no way for it to be late. Recording a moment three years after it happened produces exactly the same thing as recording it that evening.

Only Captures generated by a Commitment carry state, because only those were promised in advance:

- **Due** — generated by a Commitment, not yet answered. Lives in the inbox.
- **Kept** — the moment was recorded.
- **Missed** — the confirmation window closed with no answer (see Confirmation Window).

A Capture also carries the datetime it actually took place, and can optionally have its own color, independent of its Activity's color.

> A Capture dated August 4th, with the note "Managed to run 5km despite the drizzle" and a photo of wet shoes attached.

### Capture cost

Per Principle 7, capture is a one-tap action. Note, photos, mentions, color, and circle sharing are all optional and all deferrable. A capture saved with nothing attached is a complete, valid capture — the date and the Activity alone are a record worth keeping.

### Captures are never final

Memory is reconstructive: meaning accretes long after the fact. A Capture can be edited, annotated, and have photos added at any point in its life, including years later.

> Added in 2027 to a capture from 2024: "This was the trip right before I quit."

Editing is not versioned or flagged. A capture is the user's current understanding of a moment, not a legal record of what they thought at the time.

### Offline

Capture must work with no network. The moments most worth capturing — on a trail, mid-run, in a basement, abroad — are disproportionately the ones with no signal. A capture is written locally and synced opportunistically; the user is never blocked, and never shown a failure state for a moment they already recorded.

## Activity Commitment

The opt-in mode where an Activity stops being flexible and becomes a routine on a schedule the user sets, with the app keeping an honest record of whether they kept it.

**This mode judges the user, on purpose.** It exists for a specific and legitimate reason: *learn from your past.* Being able to look back and see plainly whether you actually showed up for something is information a warm, judgment-free archive cannot give you — and for people trying to build a practice, it is the information that matters most. A year of kept commitments is also, genuinely, a memory: the year you trained for the marathon and actually ran five days a week is part of the life story, not just a metric.

Memoria does not pretend this sits comfortably next to the rest of the app. It doesn't. It is a deliberate, bounded exception to Principle 2, and the rules below are what keep it from contaminating everything else.

### The opt-in contract

The screen where a user enables a Commitment is the most important screen in this feature, and it is written to be blunt rather than reassuring:

> **This is different from the rest of Memoria.**
>
> Memoria will generate an entry every time this activity is due, and record the days you don't answer. You will see those days when you look back at this activity.
>
> That's the point — it's how you find out whether you actually showed up. But it means this activity will sometimes tell you something you don't want to hear.
>
> You can change the schedule, loosen it, pause it, or turn it off completely at any time.

Consent is the entire mechanism that makes this mode healthy rather than corrosive. The same record, kept about a person who chose it, motivates; kept about a person who didn't, it shames. So the consent has to be real: explicit, informed, at enable time, and never buried in a settings toggle or applied by default.

### Confirmation Window

Every Capture a Commitment generates starts out **due**. The user has a configurable window (set per Commitment, with a sensible default) to answer it. If that window closes with no answer, the Capture becomes **missed**.

Missing is recorded honestly. That is the point of the mode, and softening it would be dishonest to a user who explicitly asked to be held to something.

A missed Capture can still be recorded later, at any time — being slow never locks anyone out of recording a moment that actually happened. Whether a late recording counts toward adherence depends on the Commitment's strictness.

### Strictness

The user sets how demanding a Commitment is. Strictness affects only the confirmation window and how adherence is counted — never how a moment is stored.

- **Gentle** — a wide window, and late recordings count as kept. For routines the user wants a nudge toward, not a verdict on.
- **Standard** — the default window; late recordings count as kept but are noted as late.
- **Strict** — a narrow window, and only on-time recordings count. For people who want the record to mean something.

Strictness can be changed at any time, and changing it never rewrites history.

### Adherence review

Adherence is delivered **retrospectively and in aggregate**, because that is what the mode is actually for. "Learn from your past" requires the mistake to be in the past: a notification at 9pm saying you missed today's run isn't a lesson, it's scolding about a thing you can no longer do anything about.

So the record is reported as a rate over a period, and always diagnostically rather than as a grade:

> "Morning Run, March: kept 18 of 30.
> Weekdays 85%, weekends 20%."

The second line is the valuable one. **Most missed commitments are schedule design errors, not character failures** — if someone reliably misses weekend runs, the schedule is wrong, not the person. Reporting the pattern lets them fix the schedule; reporting a bare score ("adherence: 62%") only invites them to judge themselves. Memoria always shows the pattern.

### The exit ramp

When adherence stays low, Memoria offers a way out before the user finds one themselves:

> "Morning Run: 3 of the last 20. Want to change the schedule, loosen it, or pause it for a while?"

This is the most important safeguard in the feature. If leaving a Commitment feels like admitting defeat, users don't leave the Commitment — they leave the app, and Principle 1 loses. Offering the adjustment first turns the moment of maximum shame into the moment the app is most obviously on the user's side. Pausing and loosening are presented as ordinary, sensible choices, never as giving up.

### Commitment Firewall

The rule that keeps this mode from leaking into everything else:

> **Aggregate adherence belongs in the archive. Individual failures do not.**

- **Allowed:** "2024: 210 runs," "the spring you kept this every single week," adherence rates in Patterns and Recap. Sustained discipline is part of a life story and is worth remembering.
- **Never:** a **missed** Capture appearing in Album, On This Day, Search results, or an Activity's browsable timeline. Browsing March 2024 shows the mornings you ran. It does not show a column of the ones you didn't.

An archive that ambushes you with individual failures years later is distributing shame under cover of nostalgia, and it will make the archive something the user avoids opening. A period-level record of how consistent you were does none of that — and is the thing the user opted in for in the first place.

## Capture Mention

Mentions another user who was present in or involved with that moment, on a personal (non-circle) Capture. Mentioning isn't just attribution — it's how a personal Capture, which is private by default, reaches someone outside its owner.

> "Evening run at GBK with @gede."

### How sharing works

- Mentioning a user automatically shares the Capture with them — no action needed on their part. This keeps things frictionless for passive/silent users: they simply gain a new Capture from a friend.
- If the mentioner and the mentioned user are both members of one or more of the same Circles, the app optionally offers to also share the Capture into those Circles ("Share to circle too?"), letting the user pick one or more. This step is entirely optional, never automatic.
- A Capture shared into multiple Circles at once appears in all of them simultaneously.

### Why there is no "share to circle" button

Sharing a personal Capture only ever happens as a byproduct of mentioning someone who was there. This is deliberate (Principle 5): it keeps sharing an act of *relation* rather than *publication*. The cost is real — a sunset with nobody in it cannot be shared to a circle — and that cost is accepted.

Collaborative Activities are the sanctioned exception: an Activity owned by a Circle is a shared space by construction, and captures added to it are visible to that Circle without any mention.

### Leaving a mention

A mentioned user can remove themselves from a Capture at any time. Doing so removes the mention and the Capture from their view; the Capture itself stays with its owner, unchanged apart from the mention. No notification is sent to the owner — being able to step out of someone else's record of you should not require a confrontation.

Who may mention a user at all is controlled by that user (see Privacy & Control).

## Circle

A private social circle. A place to share exclusively with close friends. Users can create collaborative activities that any member can contribute to.

> - Circle "NgodingVareng"
> - Circle "Project Team"

### Circle Invite

- Any member of a circle who has invite privileges can invite another user by entering their username, provided that user allows being invited by username.
- Any member of a circle who has invite privileges can share an invite link to the circle.
- A user can join a circle directly if invited via username. Joining via an invite link requires confirmation from a circle member who has invite privileges.

### Permissions

- The user who creates a circle becomes its admin.
- Users (other than the admin) do not have invite privileges by default.
- Users (other than the admin) can add captures to collaborative activities by default.

## Activity Schedule

Defines when a Commitment's Captures get generated. A schedule is recurring — e.g. every weekday, every Sunday morning, or a custom pattern — and one Activity can have more than one active schedule at the same time (e.g. separate "morning" and "evening" routines under the same Activity).

> "Work" recurs every weekday (Mon–Fri).
> "Evening Workout" recurs every day at 16:00.

The schedule is the mechanism; [Activity Commitment](#activity-commitment) is what it means. Editing a schedule affects future Captures only and never rewrites what was already recorded.

---

## Time

Time is Memoria's primary axis and needs to be exact.

Every Capture carries four distinct timestamps:

| Timestamp | Meaning | Editable |
|---|---|---|
| `occurred_at` | When the moment actually happened. The archive's sort and grouping key. | Yes — the user is the authority on when their life happened |
| `recorded_at` | When the user actually wrote it down. | No |
| `due_at` | For scheduled Captures, when the schedule generated it. | No |
| `captured_at` | When the user recorded it (the transition out of **open**). | No |

The gap between `occurred_at` and `recorded_at` is the raw material for Settling Time (see Patterns).

### Timezones and travel

Every Capture stores the **local UTC offset in effect where the moment happened**, alongside its instant.

The archive is browsed in the timezone of the moment, not the timezone of the viewer. A capture made in Tokyo on August 4th stays on August 4th forever, even when the user is back in Jakarta and even when they look at it in ten years. On This Day matches on that stored local calendar date.

This is not a technicality. Getting it wrong means a user's memory of a trip shifts by a day the moment they fly home, which is precisely the kind of quiet betrayal that makes an archive feel untrustworthy.

---

## Looking Back

Retrieval is where a memory app either earns its existence or fails to. Capture is the cost; looking back is the payoff. These surfaces are first-class, not reporting.

### Album

A visual gallery that aggregates all photo attachments from every Capture. A wordless visual timeline.

- **Personal Album** — every Capture belonging to the user.
- **Circle's Album** — every Capture visible within that Circle: both its own collaborative Activities and any personal Captures shared into it via mentioning. Shared-in Captures carry a label crediting their original creator (e.g., "from @ujang"), so it's clear they aren't native to the Circle.

### On This Day

A "Timehop"-style automatic flashback. Surfaces captures recorded on the same day and month in previous years.

Date matching is the crudest available relevance signal and should not be the only one. Richer resurfacing signals worth building toward:

- "You haven't looked at this in three years."
- "This is from where you are standing right now."
- "This was the last time you saw @gede."

All of it is subject to Resurfacing Controls (see Privacy & Control). Automatic resurfacing without a refusal mechanism is a well-documented way to hurt people, and Memoria will not ship the first without the second.

### Arcs

Humans remember in episodes and stretches, not in filters. An Activity is already one kind of arc ("all my runs"), but the ones that matter most cut across Activities:

> - "Everything with @gede"
> - "The year I was learning guitar"
> - "Everything at GBK"

An Arc is a saved, named, living view over the archive — defined by a person, a place, a period, or a query — that keeps collecting new captures as they arrive. This is the difference between a searchable database and a story.

### Activity Recap

An algorithmic summary of activity history over a given period, similar to a "Wrapped" feature, but for one's personal life memories. Monthly Recap (end of month) and Yearly Recap (end of year).

### Search

A global semantic/text search engine within the app, since years of accumulated memories are hard to search manually.

### Mentioned Moments

Every moment the user has ever been mentioned in.

### Patterns

Visualizations summarizing a user's routines and rhythms over time.

Every number here is framed **retrospectively**, never prospectively. "You've been running for three years" and "47-day streak" are computed from identical data and produce opposite psychology: one is identity, the other is debt. Memoria only ever renders the first.

- **Heatmap** — a GitHub-style contribution graph of Capture activity over time.
- **Settling Time** — Memoria's most distinctive stat for manual Captures. Measures how long a moment takes to settle into a record: the gap between when something happened and when the user actually wrote it down. Some moments are logged instantly because they landed hard; others take a week because their meaning only arrived later. This is texture about how memory works, **not** a punctuality score, and it must never be phrased as one. Note that this is a different measurement from Commitment adherence, which *is* about punctuality — they must not be merged or presented together.
- **Adherence** — for Commitments only: kept vs. due over a period, broken down by the dimensions that explain it (day of week, time of day, season). See [Adherence review](#adherence-review). Shown only for Activities the user put under a Commitment, and never extrapolated to manual Activities — a manual Activity has no denominator, because nothing was ever owed.
- **Activity distribution** — e.g., "This month had the most exercise activities."

**Placement matters more than design here.** These surfaces are destinations the user chooses to visit, not home-screen furniture. A heatmap glanced at monthly is a delight; the same heatmap on the home screen is a daily verdict — and it will be greyest during the exact periods of a life that are hardest, which is when a memory app must be gentlest.

### First captures

Memoria is worthless on day one and priceless in year three, which makes the first week its single largest retention risk.

**Photo backfill** closes it: on setup, the user can import from their device photo library, and existing photo timestamps become real captures immediately. A user who has never opened the app before arrives at a populated Album, a filled-in heatmap, and a working On This Day on their first session. Imported captures are ordinary captures — editable, groupable into Activities, and mentionable — not a separate second-class kind.

---

## Response

Comments and reactions are how a shared moment gets answered. They exist in three different contexts, each with its own audience:

1. **Collaborative Activity** — a Capture belonging to a circle-owned Activity. Visible to that Circle's members.
2. **Capture shared to a Circle** — a personal Capture shared into one or more Circles as part of the mention flow (see Capture Mention). Visible to each of those Circles' members.
3. **Mentioned Capture (non-circle)** — a personal Capture with mentions that wasn't shared to any Circle. Visible only to the Capture's creator and the mentioned user(s).

### Comment

A flat (no threading) contextual discussion space under a moment.

> "Oh man, I remember when our motorbike tire went flat right around here!"

### Capture Reaction

A light, private, and warm interaction toward a shared moment, from a fixed set of emoji.

> ❤️ 😂 🥹

Reaction counts are never displayed as a number — who reacted is visible, how many is not.

### Terms

- Blocked users can never view, comment, or react — regardless of any mention or circle membership they'd otherwise have. This applies both ways (see Privacy & Control).
- Muted users aren't restricted at all — their comments/reactions stay visible to everyone else. Muting just hides their activity from the muter's own view.
- When more than one of the contexts above applies to the same Capture at once, a commenter/reactor is shown a single badge indicating how they have access — "Mentioned" takes priority over a Circle-name badge, to avoid stacking multiple badges on one person.
- Visibility stays scoped to context: e.g., a mentioned user who isn't a member of a Circle the Capture was also shared to cannot see that Circle's comments/reactions, even though they can still see the Capture itself.

---

## Notification

An alert system that serves purely as a reminder. Notifications exist to bring something back to the user's attention, never to pull the user back into the app.

### Notification Types

**Commitment reminders** — only for Activities under a Commitment, and part of what the user opted into:

- **Commitment upcoming**: "'Morning Run' is coming up soon."
- **Capture due**: "How did 'Morning Run' go? If you have a photo or story, save the moment here."
- **Capture missed**: "'Morning Run' wasn't recorded today." Off by default; enabling it is a second, separate choice inside a Commitment's settings.

**Gifts** — delivered when ready:

- **On this day ready**: "A year ago today. There might be a moment you'd like to look back on."
- **Recap generated**: "Last month's colors have been wrapped up. Check out your activity recap for July!"

**Someone needs you** — delivered immediately, because a person is waiting:

- **Circle invite received**: "Gede invited you to join the circle 'NgodingVareng'."
- **Capture mentioned**: "Margarin added you to the moment 'Evening Run at GBK'."

**Responses** — **batched into a single daily digest**, never delivered individually:

- "3 people responded to your moments today."
- Expanded: reactions and comments received since the last digest.

Instant per-reaction and per-comment pings are the exact engagement mechanic Memoria exists as an alternative to. Batching preserves the information and discards the slot machine.

**On the "Capture missed" notification.** It is off by default and phrased as a bare fact — no "in time," no "again," no adverb doing moral work. The record of the miss is kept regardless; the notification only decides whether the user is told about it that evening. Most users are better served finding out in the adherence review, where it is a lesson rather than a scolding they can do nothing about. But a user who explicitly wants the daily nudge is entitled to it, and refusing them it would be Memoria substituting its judgment for theirs on the one feature they turned on specifically to be held to something.

### Notification Preferences

Users can enable or disable each type above independently, and set quiet hours during which nothing is delivered.

The defaults are the actual product decision. Memoria ships with gifts on, "someone needs you" on, the response digest on at a fixed daily hour, and Commitment reminders on for Commitments only — with "Capture missed" off until asked for. Nothing is real-time except the two cases where a human being is actually waiting for an answer.

---

## Privacy & Control

A macro-level account control system. Reinforces Memoria's position as a privacy-first app that respects data ownership.

### Social Interaction Controls

- Who can mention me
- Who can invite me to a circle
- **Block** a user — neither user can see, mention, comment on, or react to the other's Captures, in any context, in either direction.
- **Mute** a user — one-directional: their Captures, comments, and reactions are hidden from the muter's view only. The muted user isn't restricted and isn't notified; everyone else still sees their activity normally.

**Blocking never subtracts from your own archive** (Principle 4). A Capture you created that mentions a user you later block stays in your archive exactly as it was — the same photos, the same note, the same date. Only the mention itself goes inert: it stops linking to them, stops granting them access, and renders as a plain name. Losing your own memory of an evening because of how the friendship ended later would be the single worst thing this app could do.

### Resurfacing Controls

Any archive kept for a decade will eventually hold a person who died, a relationship that ended, and a period the user does not want re-entered unannounced. At Memoria's intended timescale this is a certainty, not an edge case — and these will be the app's most emotionally significant moments.

The user can permanently exclude content from all automatic resurfacing (On This Day, Recap, notifications, algorithmic Album groupings), at three granularities:

- **A single Capture**
- **Everything involving a person** — every capture mentioning them, and optionally everything from circles shared with them
- **A date range** — "don't resurface anything from 2023"

Excluded content is **not deleted and not hidden**. It stays fully in the archive, searchable and browsable, exactly where the user left it. The only thing that changes is that Memoria stops bringing it up on its own. The difference between "I can't find it" and "it won't ambush me" is the whole point.

Resurfacing controls are always reachable directly from a resurfaced item — the moment a user most needs this control is the moment they have just been shown something they did not want to see.

**Adherence data is covered by these controls too, and Memoria never pushes a decline unprompted.** A period of collapsed adherence usually marks illness, grief, burnout, or a life coming apart — the one situation where "learn from your past" is both wrong and cruel, because the lesson is not that the user should have tried harder. The data stays available when the user goes looking for it, since hiding their own history from them would be its own kind of dishonesty. But Memoria will not raise it on its own initiative, and a date range excluded from resurfacing is excluded from adherence summaries as well.

### Circle Membership

- **Visibility when leaving a circle** — leaving only affects the future: the user stops seeing new activity in that Circle and can no longer post, comment, or react there. It does not erase the past — their existing Captures shared into that Circle, and their comments/reactions, stay exactly as they were. Preserving that history is core to what Memoria is for.

### Data Controls

- Export all my data — a complete, readable archive in an open format, including photos at original quality. An archive you cannot take with you is not yours.
- Delete account
- Strip EXIF/GPS metadata from photos

---

## Lifecycle & Deletion

What survives what. These rules follow directly from Principle 4.

**A user deletes their own Capture.** It is removed everywhere: from their archive, from any Circles it was shared into, and from any mentioned users' views. Comments and reactions on it go with it. It was their moment to keep or not.

**A mentioned user removes themselves from a Capture.** The mention and their access end. The Capture is untouched for its owner. See [Leaving a mention](#leaving-a-mention).

**A Circle is dissolved.** Captures in the Circle's collaborative Activities return to the personal archives of whoever created each one. Personal Captures that were shared into the Circle simply lose that share and stay personal. Comments and reactions made in that Circle are removed along with the Circle — they were scoped to an audience that no longer exists.

**A user leaves a Circle.** Nothing they contributed is removed. See [Circle Membership](#circle-membership).

**A user blocks another user.** No content is deleted on either side. See [Social Interaction Controls](#social-interaction-controls).

**A user deletes their account.** Their own Captures, Activities, and photos are deleted. Their comments and reactions on *other people's* Captures are **anonymized, not deleted** — the discussion under someone else's moment stays coherent, attributed to a former member. Another person's departure must not tear holes in your record of a conversation you were part of.

---

## Open Decisions

Recorded here rather than silently settled, because each is a product call with real cost.

**1. Core entity naming.** "Activity" is habit-tracker vocabulary (Apple Activity rings, Strava activities) and imports the compliance framing this document works to remove. What the entity actually is — a recurring thread of a life — is better served by **Thread** or **Chapter**. Likewise, this document repeatedly slips into calling a Capture a "moment," which is a strong signal that **Moment** is the noun and *capture* is the verb.

Both renames cascade through the schema, API, web, and mobile, so neither is applied here. If the rename is not taken, the framing fix still is: Memoria should stop describing itself as an "activity-tracking app." The category a product claims determines what it gets compared against, and "activity tracker" invites a comparison with Strava that Memoria loses, while "private memory keeper" invites one it wins.

**2. ~~Whether the missed state should exist at all.~~ Settled: it stays.** Commitment mode exists precisely to keep an honest record of the days a user didn't show up, so removing the state would remove the feature. What was wrong was never the record — it was delivering it as a real-time reproach, in a word the user never consented to, on a surface meant for nostalgia. The opt-in contract, retrospective aggregate delivery, and the Commitment Firewall address each of those without blunting the mode.

**2a. Still open: does changing `confirmation_timeout_minutes` apply retroactively?** Strictness is user-adjustable, so a loosened window could either rewrite past **missed** Captures into **kept** or leave history alone. The document currently says history is never rewritten, which is the honest answer for a mode whose value depends on its record being trustworthy — but it means a user who set an unrealistically strict window early is stuck with a record that reflects a rule they've since abandoned. A middle path is recomputing adherence *rates* under current strictness while leaving individual Capture states frozen.

**3. Mention vs. Together.** "Mention" is unambiguous and already the document's working vocabulary, so it won. **Together** ("Evening run at GBK, together with @gede") is warmer and states the actual semantic — this person was there — rather than borrowing a social-media verb.

**4. Save a mentioned Capture to my own archive.** Today, when an owner deletes a Capture, it vanishes for everyone mentioned in it. That is defensible — it was the owner's moment — but it means a shared memory can be unilaterally erased from someone else's history. An explicit "keep this in my archive" action for mentioned users would resolve it, at the cost of a second copy the original owner can no longer reach.

**5. Whether the Commitment Firewall should be structural rather than a rule.** It is currently a discipline: handlers and queries are expected to exclude **missed** Captures from archive surfaces, and nothing stops a future endpoint from forgetting. Storing Commitment outcomes in their own table — separate from the Captures that represent moments that actually happened — would make the firewall impossible to breach by accident, at the cost of a more complex model. Worth deciding before the archive surfaces are built rather than after.
