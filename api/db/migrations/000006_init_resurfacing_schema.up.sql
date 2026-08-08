-- =========================================================
-- Looking Back: Lens, Echoes (resurfacing), Recap.
--
-- Resurfacing is a gift, and gifts can be declined —
-- resurfacing_exclusions is the refusal mechanism, and no
-- automatic surface may ship without consulting it.
-- =========================================================

CREATE TYPE lens_kind AS ENUM('person', 'place', 'period', 'query');

CREATE TYPE resurfacing_scope AS ENUM('moment', 'person', 'date_range');

CREATE TYPE resurfacing_reason AS ENUM('same_day', 'dormant', 'nearby', 'person');

CREATE TYPE recap_period AS ENUM('monthly', 'yearly');


-- ---------------------------------------------------------
-- lenses
-- ---------------------------------------------------------
-- A saved, named, living view over the archive, defined by a person, a
-- place, a period, or a query, that keeps collecting new Moments as
-- they arrive.
--
-- There is deliberately no lens_moments join table. A Lens is a way of
-- looking, not a container you fill — nothing is ever captured "into"
-- one. Its contents are always the result of evaluating the predicate
-- below against `moments` at read time.
CREATE TABLE lenses(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    kind lens_kind NOT NULL,
    color_hex VARCHAR(7),

    -- kind = 'person'
    person_user_id UUID REFERENCES users(id) ON DELETE CASCADE,

    -- kind = 'place'
    place_name VARCHAR(255),
    latitude NUMERIC(9, 6),
    longitude NUMERIC(9, 6),
    radius_meters INT,

    -- kind = 'period'. Local calendar dates, matched against
    -- moments.occurred_on so a Lens over "the year I was learning
    -- guitar" does not shift when the user travels.
    from_date DATE,
    to_date DATE,

    -- kind = 'query'
    query_text TEXT,

    pinned_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,

    CONSTRAINT chk_lenses_shape
    CHECK (
        (kind = 'person' AND person_user_id IS NOT NULL)
        OR (kind = 'place' AND (place_name IS NOT NULL OR latitude IS NOT NULL))
        OR (kind = 'period' AND from_date IS NOT NULL AND to_date IS NOT NULL
            AND to_date >= from_date)
        OR (kind = 'query' AND query_text IS NOT NULL)
    )
);

CREATE INDEX idx_lenses_user_id ON lenses(user_id) WHERE deleted_at IS NULL;


-- ---------------------------------------------------------
-- resurfacing_exclusions
-- ---------------------------------------------------------
-- Permanently excludes content from ALL automatic resurfacing: Echoes,
-- Recap, notifications, and algorithmic Album groupings.
--
-- Excluded content is not deleted and not hidden. It stays fully in the
-- archive, searchable and browsable, exactly where the user left it.
-- The only thing that changes is that Memoria stops bringing it up on
-- its own.
--
-- Adherence data is covered too: a 'date_range' exclusion must also be
-- applied to commitment_occurrences.due_on before any adherence summary
-- is generated or surfaced unprompted.
CREATE TABLE resurfacing_exclusions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    scope resurfacing_scope NOT NULL,

    -- scope = 'moment'
    moment_id UUID REFERENCES moments(id) ON DELETE CASCADE,

    -- scope = 'person': every Moment mentioning them, and — when
    -- include_shared_circles is set — everything from Circles shared
    -- with them as well.
    person_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    include_shared_circles BOOLEAN NOT NULL DEFAULT FALSE,

    -- scope = 'date_range', e.g. "don't resurface anything from 2023".
    -- Compared against occurred_on / due_on, i.e. stored local dates.
    from_date DATE,
    to_date DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_resurfacing_exclusions_shape
    CHECK (
        (scope = 'moment' AND moment_id IS NOT NULL
         AND person_user_id IS NULL AND from_date IS NULL AND to_date IS NULL)
        OR
        (scope = 'person' AND person_user_id IS NOT NULL
         AND moment_id IS NULL AND from_date IS NULL AND to_date IS NULL)
        OR
        (scope = 'date_range' AND from_date IS NOT NULL AND to_date IS NOT NULL
         AND to_date >= from_date AND moment_id IS NULL AND person_user_id IS NULL)
    )
);

CREATE INDEX idx_resurfacing_exclusions_user_scope
ON resurfacing_exclusions(user_id, scope);

CREATE UNIQUE INDEX uq_resurfacing_exclusions_moment
ON resurfacing_exclusions(user_id, moment_id)
WHERE scope = 'moment';

CREATE UNIQUE INDEX uq_resurfacing_exclusions_person
ON resurfacing_exclusions(user_id, person_user_id)
WHERE scope = 'person';


-- ---------------------------------------------------------
-- resurfacings
-- ---------------------------------------------------------
-- Echoes, in the schema. One row per Moment actually surfaced, so the
-- same flashback isn't served twice and so declining is always
-- reachable directly from the resurfaced item.
--
-- Date matching ('same_day') is the crudest relevance signal and
-- deliberately not the only one — 'dormant' reads moments.last_viewed_at,
-- 'nearby' reads the coordinates, 'person' reads moment_mentions.
CREATE TABLE resurfacings(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    moment_id UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
    reason resurfacing_reason NOT NULL,

    -- The user's local date this echo was served for.
    surfaced_on DATE NOT NULL,

    opened_at TIMESTAMPTZ,
    dismissed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_resurfacings_user_moment_reason_day
ON resurfacings(user_id, moment_id, reason, surfaced_on);

CREATE INDEX idx_resurfacings_user_surfaced_on
ON resurfacings(user_id, surfaced_on DESC);


-- ---------------------------------------------------------
-- recaps
-- ---------------------------------------------------------
-- Monthly and yearly period summaries. The body is JSONB because a
-- Recap is an algorithmic composition whose shape will change, and
-- because it is fully regenerable from the archive — this is a cache of
-- a rendering, not a source of truth.
--
-- Recap is an automatic resurfacing surface, so generation must honour
-- resurfacing_exclusions, including for the aggregate adherence figures
-- it is allowed to include.
CREATE TABLE recaps(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    period recap_period NOT NULL,

    -- Local calendar bounds, inclusive.
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,

    payload JSONB NOT NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    opened_at TIMESTAMPTZ,

    CONSTRAINT chk_recaps_period_bounds CHECK (period_end >= period_start)
);

CREATE UNIQUE INDEX uq_recaps_user_period_start
ON recaps(user_id, period, period_start);