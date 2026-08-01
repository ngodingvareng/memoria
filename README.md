<h1 align="center">
  Memoria
</h1>

<p align="center">
  Capture moments and history today to weave the memories of tomorrow.
</p>

## What is this dawg?

Basically, this is a full-stack application built with Golang and React (using TanStack Router). It is inspired by Instagram (for the social media side) and GitHub's commit activity (for tracking daily activities like work, traveling, etc.). What does that mean? I can't tell you just yet!

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
