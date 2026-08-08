-- =========================================================
-- The archive: Threads, Moments, their photos, and the two
-- ways a Moment reaches another person (mention, circle share).
--
-- This is the whole product. Note what is NOT here: `moments`
-- has no status column. Commitment outcomes (due/kept/missed)
-- live in commitment_occurrences, one migration later. A
-- `missed` day produces no row in this file at all, which is
-- what makes the Commitment Firewall structural instead of a
-- rule every future query has to remember.
-- =========================================================

-- Provenance only. This never gates access, never changes how a Moment
-- is stored, and is never used to implement the Firewall (that is the
-- absence of a commitment_occurrences row, not this column). Imported
-- Moments are ordinary Moments — editable, groupable into Threads, and
-- mentionable — exactly like manual ones. Set once at insert and never
-- updated, so unlike the old threads.has_commitment flag it cannot
-- drift out of sync with anything.
CREATE TYPE moment_origin AS ENUM('manual', 'commitment', 'import');


-- ---------------------------------------------------------
-- threads
-- ---------------------------------------------------------
-- A recurring strand of a life, and the container for its Moments.
--
-- user_id is always the creator and, for a personal Thread, its owner.
-- When circle_id is set the Thread is collaborative and owned by the
-- Circle: access is governed by circle_members, not by user_id, and any
-- member with can_capture may add Moments to it.
--
-- Deliberately absent:
--   * has_commitment — derived from commitments (paused_at/archived_at
--     IS NULL), removing the drift hazard logged as SCHEMA_REVIEW #5.
--   * confirmation_timeout_minutes — moved to commitments, where the
--     spec puts it: two Commitments under one Thread (a strict morning
--     run, a gentle evening walk) must be able to differ.
CREATE TABLE threads(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- CASCADE, not SET NULL: dissolving a Circle removes the shared
    -- container. Its Moments survive via moments.thread_id ON DELETE
    -- SET NULL and return to the personal archive of whoever captured
    -- each one.
    circle_id UUID REFERENCES circles(id) ON DELETE CASCADE,

    name VARCHAR(255) NOT NULL,
    description TEXT,
    color_hex VARCHAR(7),

    -- Manual ordering on the Thread list; the archive itself is always
    -- ordered by occurred_at.
    sort_order INT NOT NULL DEFAULT 0,
    archived_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_threads_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$')
);

CREATE INDEX idx_threads_user_id ON threads(user_id) WHERE deleted_at IS NULL;

CREATE INDEX idx_threads_circle_id
ON threads(circle_id)
WHERE circle_id IS NOT NULL AND deleted_at IS NULL;


-- ---------------------------------------------------------
-- thread_images
-- ---------------------------------------------------------
-- Images describing the Thread itself (cover photo / gallery), as
-- opposed to moment_images, which are photos of a specific occurrence.
-- Only moment_images feed the Album.
CREATE TABLE thread_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    thread_id UUID NOT NULL REFERENCES threads(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_thread_images_thread_id ON thread_images(thread_id, sort_order);


-- ---------------------------------------------------------
-- moments
-- ---------------------------------------------------------
CREATE TABLE moments(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- The capturer, and the row's real owner. Distinct from
    -- threads.user_id because a collaborative Thread holds Moments from
    -- several people, and because the lifecycle rule "Moments return to
    -- the personal archives of whoever captured each one" needs a
    -- per-Moment answer that survives the Thread disappearing.
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Nullable, and SET NULL rather than CASCADE. A Moment can exist
    -- before it has any parent (photo backfill creates Moments with no
    -- Thread), and must outlive its container when a Circle dissolves.
    -- Hard-deleting a *personal* Thread that should take its Moments
    -- with it is therefore an explicit two-statement operation in the
    -- usecase layer, not a silent cascade.
    thread_id UUID REFERENCES threads(id) ON DELETE SET NULL,

    origin moment_origin NOT NULL DEFAULT 'manual',

    -- ----- The four timestamps, per the spec's Time table -----

    -- When it actually happened. The archive's sort and grouping key,
    -- and the only one of the four the user may edit.
    occurred_at TIMESTAMPTZ NOT NULL,

    -- The same instant as wall-clock time where it happened, plus the
    -- UTC offset that was in effect there. Stored rather than derived
    -- because the viewer's zone must never influence the archive: a
    -- Moment captured in Tokyo on August 4th stays on August 4th
    -- forever, including ten years later from Jakarta.
    --
    -- Invariant maintained by the write path (not expressible as a
    -- CHECK, since timestamptz -> timestamp conversion is STABLE, not
    -- IMMUTABLE):
    --     occurred_local = occurred_at AT TIME ZONE 'UTC'
    --                      + occurred_utc_offset_minutes
    occurred_local TIMESTAMP NOT NULL,
    occurred_utc_offset_minutes SMALLINT NOT NULL DEFAULT 0,

    -- The stored local calendar date. This — never occurred_at — is
    -- what Echoes match on and what the Rhythms heatmap buckets by.
    -- Generated so the two can never disagree.
    occurred_on DATE GENERATED ALWAYS AS (occurred_local::date) STORED,

    -- When the user actually wrote it down. Never editable. For offline
    -- capture the client supplies the true local record time at sync,
    -- which is why this is not simply DEFAULT NOW().
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Time to Tell, materialized. Meaningful only for Moments with no
    -- commitment_occurrences row — the spec is explicit that Time to
    -- Tell is a manual-Moment stat and must never be merged with, or
    -- presented alongside, Commitment adherence.
    settling_time INTERVAL GENERATED ALWAYS AS (recorded_at - occurred_at) STORED,

    -- due_at and captured_at are deliberately NOT columns here. They
    -- describe a schedule obligation, not something that happened, and
    -- live on commitment_occurrences.

    -- ----- Content, all optional -----
    -- A Moment saved with nothing attached is a complete, valid Moment:
    -- the date alone is a record worth keeping (Principle 7).
    note TEXT,
    color_hex VARCHAR(7),

    -- Location, for the "from where you are standing right now" Echo
    -- and for place-defined Lenses. Plain columns, no PostGIS: bounded
    -- radius queries over one user's archive don't justify the
    -- extension.
    place_name VARCHAR(255),
    latitude NUMERIC(9, 6),
    longitude NUMERIC(9, 6),

    -- ----- Search -----
    -- 'simple' rather than 'english': notes are written in whatever
    -- language the user thinks in, and stemming the wrong one is worse
    -- than not stemming at all. Semantic search is a later layer over
    -- moment_embeddings (see the commented block at the end of this
    -- file), not a replacement for this.
    search_document tsvector GENERATED ALWAYS AS (
        to_tsvector(
            'simple'::regconfig,
            coalesce(note, '') || ' ' || coalesce(place_name, '')
        )
    ) STORED,

    -- ----- Offline capture -----
    -- Client-generated id, so a Moment written with no network and
    -- synced later cannot be inserted twice. Capture must never show
    -- the user a failure state for something they already recorded.
    client_id UUID,

    -- Owner-only. Feeds the "you haven't looked at this in three years"
    -- Echo signal. Updated *only* when the viewer is moments.user_id:
    -- this is not a read receipt and not a view count, and is never
    -- exposed to anyone but the owner.
    last_viewed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_moments_color_hex
    CHECK (color_hex ~ '^#[0-9A-Fa-f]{6}$'),

    -- ±17h covers every real zone including historical outliers.
    CONSTRAINT chk_moments_utc_offset_range
    CHECK (occurred_utc_offset_minutes BETWEEN -1020 AND 1020),

    CONSTRAINT chk_moments_coordinates
    CHECK (
        (latitude IS NULL AND longitude IS NULL)
        OR (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
);

-- Personal archive timeline, Album, Search, and Recap all scan one
-- user's Moments newest-first.
CREATE INDEX idx_moments_user_occurred_at
ON moments(user_id, occurred_at DESC)
WHERE deleted_at IS NULL;

-- A Thread's browsable timeline.
CREATE INDEX idx_moments_thread_occurred_at
ON moments(thread_id, occurred_at DESC)
WHERE thread_id IS NOT NULL AND deleted_at IS NULL;

-- Serves both the Rhythms heatmap (bucket by local date) and Echoes
-- (same-day-and-month in previous years, expressed as an IN-list of
-- prior-year dates rather than an expression index).
CREATE INDEX idx_moments_user_occurred_on
ON moments(user_id, occurred_on)
WHERE deleted_at IS NULL;

CREATE INDEX idx_moments_search_document
ON moments USING GIN(search_document);

CREATE INDEX idx_moments_location
ON moments(latitude, longitude)
WHERE latitude IS NOT NULL AND deleted_at IS NULL;

-- Idempotent offline sync.
CREATE UNIQUE INDEX uq_moments_user_client_id
ON moments(user_id, client_id)
WHERE client_id IS NOT NULL;


-- ---------------------------------------------------------
-- moment_images
-- ---------------------------------------------------------
-- The Album is exactly this table joined back to moments for ordering
-- and visibility. Because a missed Commitment produces no Moment, it
-- can produce no photo either — the Firewall holds here for free.
CREATE TABLE moment_images(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    image_path TEXT NOT NULL,
    image_alt VARCHAR(255),
    content_type VARCHAR(100),
    byte_size BIGINT,
    width INT,
    height INT,

    -- Whether EXIF/GPS was actually stripped for this file. Recorded
    -- per file rather than read from users.strip_photo_metadata, so
    -- flipping the setting later never misreports what was already
    -- uploaded.
    metadata_stripped BOOLEAN NOT NULL DEFAULT FALSE,

    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_moment_images_moment_id ON moment_images(moment_id, sort_order);


-- ---------------------------------------------------------
-- moment_mentions
-- ---------------------------------------------------------
-- Naming another person present. Mentioning is the only way a personal
-- Moment reaches someone outside its owner, and it grants access
-- automatically with no action needed from the mentioned user.
CREATE TABLE moment_mentions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,

    -- SET NULL, not CASCADE. When the mentioned user deletes their
    -- account the mention goes inert but the owner's Moment is
    -- otherwise untouched — same photos, same note, same date
    -- (Principle 4). It then renders from display_name below.
    mentioned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Name snapshot taken at mention time. This is what "renders as a
    -- plain name" means when the link goes inert — whether from account
    -- deletion (above) or from a block, which is evaluated live against
    -- user_blocks and deliberately not stored, so unblocking restores
    -- the link on its own.
    display_name VARCHAR(255) NOT NULL,

    -- The mentioned user removed themselves. The mention and their
    -- access end; the Moment stays with its owner, unchanged apart from
    -- the mention. No notification is sent to the owner.
    removed_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_moment_mentions_moment_user
ON moment_mentions(moment_id, mentioned_user_id)
WHERE mentioned_user_id IS NOT NULL AND removed_at IS NULL;

-- "Mentioned Moments" surface, and the person-signal for Echoes
-- ("this was the last time you saw @gede").
CREATE INDEX idx_moment_mentions_mentioned_user_id
ON moment_mentions(mentioned_user_id)
WHERE mentioned_user_id IS NOT NULL AND removed_at IS NULL;


-- ---------------------------------------------------------
-- moment_circles
-- ---------------------------------------------------------
-- A personal Moment shared into one or more Circles as an optional
-- byproduct of mentioning someone. There is no "share to circle"
-- button, so a row here always originates from the mention flow.
-- A Moment shared into several Circles appears in all of them at once.
--
-- Moments captured *into* a collaborative Thread are NOT represented
-- here: they are visible to that Circle by construction, through
-- threads.circle_id.
--
-- CASCADE on circle_id is the lifecycle rule: dissolving a Circle makes
-- a shared personal Moment simply lose that share and stay personal.
CREATE TABLE moment_circles(
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    circle_id UUID NOT NULL REFERENCES circles(id) ON DELETE CASCADE,
    shared_by_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    shared_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY(moment_id, circle_id)
);

-- A Circle's Album and activity feed.
CREATE INDEX idx_moment_circles_circle_id ON moment_circles(circle_id, shared_at DESC);


-- ---------------------------------------------------------
-- moment_embeddings (deferred — requires the pgvector extension)
-- ---------------------------------------------------------
-- Semantic Search, filled by the ai/ service. Left commented out so the
-- schema stays runnable on a stock Postgres image; uncomment together
-- with adding pgvector to docker-compose.dev.yaml. Dimension must match
-- whichever embedding model ai/ settles on.
--
-- CREATE EXTENSION IF NOT EXISTS vector;
--
-- CREATE TABLE moment_embeddings(
--     moment_id UUID PRIMARY KEY REFERENCES moments(id) ON DELETE CASCADE,
--     embedding vector(1536) NOT NULL,
--     model VARCHAR(100) NOT NULL,
--     source_hash VARCHAR(64) NOT NULL,   -- skip re-embedding unchanged notes
--     created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
-- );
--
-- CREATE INDEX idx_moment_embeddings_vector
-- ON moment_embeddings USING hnsw (embedding vector_cosine_ops);