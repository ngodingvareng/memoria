# Memoria Features

## Activity

The main entity that defines an activity, hobby, or routine. It acts as a container for specific moments (Captures). An Activity can be personal (owned by a single user) or collaborative, owned by a Circle (see the Circle section).

> My Activities:
>
> - "Morning Run"
> - "Coding Together"
> - "Coffee Exploring"

## Activity Capture

A journal entry recorded for a single moment within an Activity. This is the core of Memoria.

> A note dated August 4th about "Managed to run 5km despite the drizzle," with a photo of wet shoes attached.

## Capture Tag

Mentions another user who was present in or involved with that moment.

> "Evening run at GBK with @gede."

### Terms

- A user who mentions another user in their own personal capture automatically shares it with the mentioned user.
- A user who mentions another user in their own personal capture can optionally share it with any Circle where the mentioned user is a member.

## Circle

A private social circle. A place to share exclusively with close friends. Users can create collaborative activities that any member can contribute to.

> Circle "NgodingVareng"
> Circle "Project Team"

### Circle Invite

- Any member of a circle who has invite privileges can invite another user by entering their username, provided that user allows being invited by username.
- Any member of a circle who has invite privileges can share an invite link to the circle.
- A user can join a circle directly if invited via username. Joining via an invite link requires confirmation from a circle member who has invite privileges.

### Permissions

- The user who creates a circle becomes its admin.
- Users (other than the admin) do not have invite privileges by default.
- Users (other than the admin) can add captures to collaborative activities by default.

## Activity Schedule

Scheduling of planned future activities. Used to trigger reminders so users don't forget to log their moment after the activity is done.

> Recurring (e.g., every Sunday morning).

## Statistic

Visualization summarizing a user's routine and habit data over time.

> - Heatmap graph (similar to GitHub's)
> - Activity distribution chart (e.g., "This month had the most exercise activities")

## Capture Reaction in the Circle

A light, private, and warm interaction from circle members or mentioned users toward a shared moment.

> ❤️ 😂 🥹

## Comment in the Circle and Shared Capture

A specific, contextual discussion space under a moment, visible only to circle members or mentioned users.

> "Oh man, I remember when our motorbike tire went flat right around here!"

### Terms

- A commenter gets a "Mentioned" badge if they are the person who was mentioned.
- A commenter gets a badge with the group's name if they are a member of the circle.
  If a commenter is both mentioned and a circle member, the "Mentioned" badge takes priority and is shown.

## Notification

An alert system that serves purely as a reminder.

### Notification Types

- **On this day ready**: "A year ago today. There might be a moment you'd like to look back on."
- **Recap generated**: "Last month's colors have been wrapped up. Check out your activity recap for July!"
- **Circle invite received**: "Gede invited you to join the circle 'NgodingVareng'."
- **Capture tagged**: "Margarin added you to the moment 'Evening Run at GBK'."
- **Reaction added**: "Margarin reacted ❤️ to your graduation moment."
- **Comment added**: "Gede commented: 'what is this'"
- **Schedule upcoming**: "The activity 'Morning Run' is coming up soon. Don't forget."
- **Capture awaiting**: "How did 'Morning Run' go? If you have a photo or story, save the moment here."
- **Capture missed**: "'Morning Run' today wasn't logged in time."

### Notification Preferences

Users can configure which notification types they want enabled.

## Album

A visual gallery that aggregates all photo attachments from every Capture. A wordless visual timeline.

> - Personal Album
> - Circle's Album

## Capture On This Day

A "Timehop"-style automatic flashback feature. Surfaces captures recorded on the same day and month in previous years.

## Activity Recap

An algorithmic summary of activity history over a given period, similar to a "Wrapped" feature, but for one's personal life memories. Monthly Recap (end of month) and Yearly Recap (end of year).

## Mentions personal view

See every moment I've ever been tagged in.

## Search Anything

A global semantic/text search engine within the app, since years of accumulated memories are hard to search manually.

## User Privacy

A macro-level account control system. Reinforces Memoria's position as a privacy-first app that respects data ownership.

### Social Interaction Controls

- Who can mention/tag me
- Who can invite me to a circle
- Block/mute specific users

### Identity Controls

- Visibility when leaving a circle

### Data Controls

- Export all my data
- Delete account
- Strip EXIF/GPS metadata from photos
