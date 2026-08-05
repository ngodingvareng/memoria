<h1 align="center">
  Memoria
</h1>

<p align="center">
  Capture moments and history today to weave the memories of tomorrow.
</p>

## What is this dawg?

Memoria is a full-stack activity-tracking app built with Golang and React (using TanStack Router), built around a simple philosophy: capture your moments and history today so you can look back and reminisce later. It's inspired by Instagram (for the social side) and GitHub's contribution graph (for visualizing daily activity over time).

The core entity is the **Activity** — e.g. "Work," "Morning Run," "Reading." Each Activity can have many **Activity Items**, its individual occurrences (one Activity to many Items). An Activity can either follow a **fixed schedule** (e.g. "Work" recurs every weekday) or have items added manually whenever the user wants.

When an Activity has a fixed schedule, Memoria automatically generates its Items as they become due. Each Item starts as "awaiting confirmation" until the user checks it off as done or not done, optionally attaching a note (text + multiple photos) about what happened. If an Item isn't confirmed within a configurable timeout (with a sensible default) after it appears, it automatically flips to "not done."

Every Item also carries its own scheduled datetime and can be given a color, letting Activities and Items be told apart visually. On top of that, Memoria includes a statistics view: a GitHub-style contribution heatmap of your activity history, plus a chart showing the confirmation delay for fixed-schedule items — how long it typically takes between an Item appearing and you actually confirming it.

See [`FEATURES.md`](./FEATURES.md) for the full feature breakdown, including Circles (private social groups), Captures, mentions, notifications, albums, and recaps.

## Get started

### Requirements

- Linux OS (As seen in api/Makefile and the commands below, some shell scripts rely on Linux commands. Cross-platform support will be added later).
- Go 1.26
- Make (If you don't have make installed, you can copy the commands directly from api/Makefile and run them manually in your terminal).
- Docker (Required for the containerized database and running integration tests).
- Bun (Recommended. Alternatively, you can use Node.js with npm, pnpm, or yarn).

### Download repository

Just clone the repository.

```sh
git clone https://github.com/ngodingvareng/memoria.git

// or

gh repo clone ngodingvareng/memoria
```

### Install packages/dependencies

Install both of the API and the WEB application.

```sh
(cd api && cp .env.example .env)
(cd api && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
(cd api && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
(cd api && go install github.com/swaggo/swag/cmd/swag@latest)
(cd api && go install github.com/vektra/mockery/v3@v3.7.2)
(cd api && go mod tidy)
(cd web && bun install)
```

### Run the application

Run both the API and the WEB application. Starting the API will automatically spin up the database container via Docker (you can check the Makefile to see how it works)."

```sh
(cd api && make dev)
(cd web && bun dev)
```

## Architecture

### Backend

We use clean architecture for this app stack. We also use unit of work for repository pattern (see https://rednafi.com/go/repo-txn-uow)

- Techstack: Golang, Fiber, SQLC, Postgres, Mockery, Golang migrate, Swaggo, RustFS

### Frontend

- Techstack: React, Tanstack Router, Tanstack Form, Bun, Vite, Tailwindcss, ShadCN
