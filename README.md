<h1 align="center">
  Memoria
</h1>

<p align="center">
  Capture moments and history today to weave the memories of tomorrow.
</p>

## What is this dawg?

Basically this is a fullstack application with Golang and React (using Tanstack router). I inspired from Instagram (as social media) and GitHub commit activity (as activity capturing. Any activities, like work, travelling, etc). What is that mean? I can't tell you right now.

## Get started

### Requirements

- You are using Linux (as you can see in `api/Makefile` and some commands below, there are some commands using linux commands, I'll fix them later so it can be run in everywhere)
- Golang 1.26
- Make (if you don't have this thing, you can copy the commands from the `api/Makefile` and run manually in your terminal)
- Docker (Used for containerized database and integration test)
- Bun (recommended, alternatively you can use Nodejs with npm/pnpm/yarn or anything you want)

### Download repository

Just clone the repo

```sh
git clone https://github.com/ngodingvareng/memoria.git

// or

gh repo clone ngodingvareng/memoria
```

### Install packages/dependencies

Install in both of api and web

```sh
(cd api && go mod tidy)
(cd web && bun install)
```

### Run you application

Run both apps. The API will also start the database container (just look at Makefile if want to know how it works).

```sh
(cd api && make dev)
(cd web && bun dev)
```
