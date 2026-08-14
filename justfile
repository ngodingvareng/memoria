# Convenience wrapper only — each project still has its own toolchain and
# lifecycle (see CLAUDE.md). Run from the repo root.

# Run the API and the web app dev servers together (Ctrl+C stops both)
[parallel]
dev: api-dev web-dev

# Run only the API dev server
@api-dev:
    cd api && just dev

# Run only the web app dev server
@web-dev:
    cd web && bun dev
