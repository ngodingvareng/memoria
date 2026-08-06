<h1 align="center">
  Memoria
</h1>

<p align="center">
  Capture moments and history today to weave the memories of tomorrow.
</p>

## What is this dawg?

Memoria is a full-stack private memory keeper built with Golang and React (using TanStack Router), built around a simple philosophy: capture your moments and history today so you can look back and reminisce later. It's inspired by Instagram (for the social side) and GitHub's contribution graph (for visualizing daily activity over time).

Its value is deferred by design — almost nothing it does matters on day one, and all of it compounds by year three. The archive is the product.

The core entity is the **Activity** — e.g. "Work," "Morning Run," "Reading" — a recurring thread of a life rather than a task. Each Activity can have many **Captures**, its individual occurrences (one Activity to many Captures). An Activity can either follow a **fixed schedule** (e.g. "Work" recurs every weekday) or have Captures added manually whenever the user wants, which is the primary way Memoria is meant to be used.

Manual Captures have no state at all — they are created and they exist. Putting an Activity under a fixed schedule turns it into a **Commitment**, an opt-in mode where Memoria generates a Capture each time it comes due (**due** → **kept**, or **missed** if the confirmation window closes unanswered) and keeps an honest record of whether the user showed up. Commitment mode deliberately judges the user; that is what it's for, so that you can look back and learn from what you actually did. It's explained bluntly at opt-in, its strictness is user-set, and the record it produces reaches the archive only in aggregate — a year's adherence rate is a memory worth keeping, a wall of individual missed days is not.

Every Capture also carries the datetime it actually took place and can be given a color, letting Activities and Captures be told apart visually. On top of that, Memoria includes a patterns view: a GitHub-style contribution heatmap of your activity history, plus **Settling Time** — how long a moment takes to settle into a record, i.e. the gap between something happening and you actually writing it down. It's texture about how memory works, not a punctuality score.

See [`FEATURES.md`](./FEATURES.md) for the full feature breakdown — including the principles and non-goals every feature is checked against, plus Circles (private social groups), mentions, notifications, albums, and recaps.

This is a monorepo with four projects, each with its own toolchain:

- **`api/`** — the Go backend. The most mature and actively developed part of the stack.
- **`web/`** — the React web app.
- **`mobile/`** — the Flutter mobile app. Early scaffold, not yet wired to the API.
- **`ai/`** — a FastAPI service meant to be called by the API backend to power future AI-driven features (e.g. activity summarizers, recaps). Still a bare scaffold; real implementation is well down the road.

## Get started

### Requirements

- Linux OS (As seen in api/Makefile and the commands below, some shell scripts rely on Linux commands. Cross-platform support will be added later).
- Go 1.26
- Make (If you don't have make installed, you can copy the commands directly from api/Makefile and run them manually in your terminal).
- Docker (Required for the containerized database and running integration tests).
- Bun (Recommended. Alternatively, you can use Node.js with npm, pnpm, or yarn).
- Flutter (stable channel) — only needed if you're working on `mobile/`.
- Python 3.14+ and [uv](https://docs.astral.sh/uv/) — only needed if you're working on `ai/`.

### Download repository

Just clone the repository.

```sh
git clone https://github.com/ngodingvareng/memoria.git

// or

gh repo clone ngodingvareng/memoria
```

### Install packages/dependencies

Install dependencies for whichever projects you plan to work on. `api/` and `web/` are the actively developed ones; `mobile/` and `ai/` are early scaffolds, so only set those up if you're touching them.

```sh
# api
(cd api && cp .env.example .env)
(cd api && go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest)
(cd api && go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest)
(cd api && go install github.com/swaggo/swag/cmd/swag@latest)
(cd api && go install github.com/vektra/mockery/v3@v3.7.2)
(cd api && go mod tidy)

# web
(cd web && bun install)

# mobile (optional, only if you're working on the Flutter app)
(cd mobile && flutter pub get)

# ai (optional, only if you're working on the AI service)
(cd ai && uv sync)
```

### Run the application

Run the API and the web app. Starting the API will automatically spin up the database container via Docker (you can check the Makefile to see how it works).

```sh
(cd api && make dev)
(cd web && bun dev)
```

To run `mobile/` or `ai/` as well:

```sh
(cd mobile && flutter run)
(cd ai && uv run fastapi dev main.py)
```

## Architecture

### Backend

We use clean architecture for this app stack. We also use unit of work for repository pattern (see https://rednafi.com/go/repo-txn-uow)

- Techstack: Golang, Fiber, SQLC, Postgres, Mockery, Golang migrate, Swaggo, RustFS

### Frontend

- Techstack: React, Tanstack Router, Tanstack Form, Bun, Vite, Tailwindcss, ShadCN

### Mobile

- Techstack: Flutter, Dart

### AI Service

- Techstack: Python, FastAPI, uv
