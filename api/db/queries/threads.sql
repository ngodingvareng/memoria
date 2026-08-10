-- name: CreateThread :one
-- circle_id NULL creates a personal Thread owned by user_id; setting it
-- creates a collaborative Thread owned by the Circle instead (see the
-- header comment on the threads table).
INSERT INTO threads(user_id, circle_id, name, description, color_hex)
VALUES (
    sqlc.arg(user_id), sqlc.narg(circle_id), sqlc.arg(name),
    sqlc.narg(description), sqlc.narg(color_hex)
)
RETURNING *;

-- name: GetThreadByID :one
-- No ownership/membership filter — callers on a collaborative Thread
-- must authorize separately (see GetThreadWithAccess) since access
-- there is governed by circle_members, not user_id.
SELECT *
FROM threads
WHERE id = sqlc.arg(id)
    AND deleted_at IS NULL;

-- name: GetThreadWithAccess :one
-- "Can this viewer reach this Thread at all" — true for the Thread's
-- creator, and for any active member of the Circle that owns it when
-- it's collaborative.
SELECT threads.*
FROM threads
WHERE threads.id = sqlc.arg(id)
    AND threads.deleted_at IS NULL
    AND (
        threads.user_id = sqlc.arg(user_id)
        OR threads.circle_id IN (
            SELECT circle_id
            FROM circle_members
            WHERE user_id = sqlc.arg(user_id)
                AND left_at IS NULL
        )
    );

-- name: ListPersonalThreadsByUserID :many
-- The "My Threads" list: Threads this user created that are not owned
-- by a Circle.
SELECT *
FROM threads
WHERE user_id = sqlc.arg(user_id)
    AND circle_id IS NULL
    AND deleted_at IS NULL
ORDER BY sort_order, created_at DESC;

-- name: ListThreadsByCircleID :many
-- A Circle's collaborative Threads.
SELECT *
FROM threads
WHERE circle_id = sqlc.arg(circle_id)
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SearchThreads :many
-- Search & filter the authenticated user's thread list: their own
-- personal Threads, plus every collaborative Thread owned by a Circle
-- they're an active member of (mirrors GetThreadWithAccess's notion of
-- "reachable"). All filters are optional (NULL = "don't filter on this"):
--   - name: fuzzy/typo-tolerant match against the thread's name and
--     description (search_document) — plainto_tsquery for word/stem
--     matching, OR pg_trgm similarity for typos, superseding the old
--     ILIKE '%...%' partial-substring-only match.
--   - archived: exact match against (archived_at IS NOT NULL)
-- Ordered by sort_order then newest-first by default; passing
-- sort_by_recency = true switches to updated_at DESC instead (backs the
-- moment-creation Thread picker's "most recently updated" suggestions).
-- The CASE evaluates to NULL for every row when false, so all rows tie
-- on that key and the ordering falls through to sort_order/created_at
-- unchanged — same query, same default behavior for every other caller.
-- Paginated via page_limit / page_offset.
SELECT *
FROM threads
WHERE (
        (threads.user_id = sqlc.arg(user_id) AND threads.circle_id IS NULL)
        OR threads.circle_id IN (
            SELECT circle_id
            FROM circle_members
            WHERE circle_members.user_id = sqlc.arg(user_id)
                AND left_at IS NULL
        )
    )
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR search_document @@ plainto_tsquery('simple', sqlc.narg(name)::text)
        OR name % sqlc.narg(name)::text
    )
    AND (
        sqlc.narg(archived)::bool IS NULL
        OR (archived_at IS NOT NULL) = sqlc.narg(archived)::bool
    )
ORDER BY
    CASE WHEN sqlc.arg(sort_by_recency)::bool THEN updated_at END DESC NULLS LAST,
    sort_order,
    created_at DESC
LIMIT sqlc.arg(page_limit)
OFFSET sqlc.arg(page_offset);

-- name: CountSearchThreads :one
-- Total matching row count for SearchThreads, ignoring
-- page_limit/page_offset — used to compute pagination metadata.
SELECT COUNT(*)
FROM threads
WHERE (
        (threads.user_id = sqlc.arg(user_id) AND threads.circle_id IS NULL)
        OR threads.circle_id IN (
            SELECT circle_id
            FROM circle_members
            WHERE circle_members.user_id = sqlc.arg(user_id)
                AND left_at IS NULL
        )
    )
    AND deleted_at IS NULL
    AND (
        sqlc.narg(name)::text IS NULL
        OR search_document @@ plainto_tsquery('simple', sqlc.narg(name)::text)
        OR name % sqlc.narg(name)::text
    )
    AND (
        sqlc.narg(archived)::bool IS NULL
        OR (archived_at IS NOT NULL) = sqlc.narg(archived)::bool
    );

-- name: SearchThreadSuggestions :many
-- Top-N thread matches for the live-typing search popover
-- (GET /search/suggestions). Same "reachable" scope as SearchThreads,
-- no archived/name filter args, no pagination — a fixed-size list.
-- prefix_query is a caller-built ':*'-suffixed tsquery string (see
-- usecase.buildPrefixTsQuery); query is the raw trimmed search text,
-- used for the trigram fallback/ranking.
SELECT threads.*
FROM threads
WHERE (
        (threads.user_id = sqlc.arg(user_id) AND threads.circle_id IS NULL)
        OR threads.circle_id IN (
            SELECT circle_id
            FROM circle_members
            WHERE circle_members.user_id = sqlc.arg(user_id)
                AND left_at IS NULL
        )
    )
    AND deleted_at IS NULL
    AND (
        search_document @@ to_tsquery('simple', sqlc.arg(prefix_query))
        OR name % sqlc.arg(query)::text
    )
ORDER BY
    ts_rank(search_document, to_tsquery('simple', sqlc.arg(prefix_query))) DESC,
    similarity(name, sqlc.arg(query)::text) DESC,
    updated_at DESC
LIMIT sqlc.arg(limit_count);

-- name: UpdateThread :one
UPDATE threads
SET name = sqlc.arg(name),
    description = sqlc.narg(description),
    color_hex = sqlc.narg(color_hex),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL
    RETURNING *;

-- name: UpdateThreadSortOrder :exec
-- Manual reordering on the Thread list; the archive itself always
-- stays ordered by occurred_at regardless of this value.
UPDATE threads
SET sort_order = sqlc.arg(sort_order),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: ArchiveThread :exec
UPDATE threads
SET archived_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: UnarchiveThread :exec
UPDATE threads
SET archived_at = NULL,
    updated_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: SoftDeleteThread :exec
UPDATE threads
SET deleted_at = NOW()
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id)
    AND deleted_at IS NULL;

-- name: RestoreThread :exec
UPDATE threads
SET deleted_at = NULL
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);

-- name: HardDeleteThread :exec
DELETE FROM threads
WHERE id = sqlc.arg(id)
    AND user_id = sqlc.arg(user_id);
