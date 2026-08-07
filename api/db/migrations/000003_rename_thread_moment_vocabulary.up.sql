-- Renames the activity-era schema to the product vocabulary defined in
-- FEATURES.md ("Vocabulary" section):
--
--   Activity         -> Thread      a recurring strand of a life
--   ActivityCapture  -> Moment      a single occurrence within a Thread
--   ActivitySchedule -> Commitment  opt-in scheduled accountability
--
-- Capture is a verb from here on, never a noun. Child entities drop the
-- parent prefix where their own noun stands alone (see CODING_STANDARDS
-- §3.1), so activity_capture_images becomes moment_images rather than
-- thread_moment_images.
--
-- Moment states are renamed with the same pass: 'awaiting' -> 'due' and
-- 'captured' -> 'kept'. 'missed' is deliberately unchanged — under an
-- opt-in Commitment the honest word is the right one.
--
-- This is a pure rename. No column is added, dropped, or retyped, and no
-- row is rewritten.

-- ---------------------------------------------------------
-- Enum
-- ---------------------------------------------------------
ALTER TYPE activity_capture_status RENAME TO moment_status;

ALTER TYPE moment_status RENAME VALUE 'awaiting' TO 'due';
ALTER TYPE moment_status RENAME VALUE 'captured' TO 'kept';

-- ---------------------------------------------------------
-- Tables
-- ---------------------------------------------------------
ALTER TABLE activities                  RENAME TO threads;
ALTER TABLE activity_images             RENAME TO thread_images;
ALTER TABLE activity_schedules          RENAME TO commitments;
ALTER TABLE activity_schedule_histories RENAME TO commitment_histories;
ALTER TABLE activity_captures           RENAME TO moments;
ALTER TABLE activity_capture_images     RENAME TO moment_images;

-- ---------------------------------------------------------
-- Columns
-- ---------------------------------------------------------
-- "fixed schedule" is no longer product vocabulary; the flag means "this
-- Thread carries at least one Commitment".
ALTER TABLE threads RENAME COLUMN is_fixed_schedule TO has_commitment;

ALTER TABLE thread_images RENAME COLUMN activity_id TO thread_id;

ALTER TABLE commitments RENAME COLUMN activity_id TO thread_id;

ALTER TABLE commitment_histories RENAME COLUMN activity_id TO thread_id;
ALTER TABLE commitment_histories RENAME COLUMN schedule_id TO commitment_id;

ALTER TABLE moments RENAME COLUMN activity_id  TO thread_id;
ALTER TABLE moments RENAME COLUMN schedule_id  TO commitment_id;
-- Matches the Time table in FEATURES.md: due_at is when a Commitment
-- generated the Moment, captured_at is when the user recorded it.
ALTER TABLE moments RENAME COLUMN scheduled_at TO due_at;
ALTER TABLE moments RENAME COLUMN confirmed_at TO captured_at;

ALTER TABLE moment_images RENAME COLUMN activity_capture_id TO moment_id;

-- ---------------------------------------------------------
-- Constraints
-- ---------------------------------------------------------
-- Renaming a constraint also renames its backing index, so primary-key
-- and unique constraints are handled here rather than under Indexes.
ALTER TABLE threads RENAME CONSTRAINT activities_pkey                  TO threads_pkey;
ALTER TABLE threads RENAME CONSTRAINT activities_user_id_fkey          TO threads_user_id_fkey;
ALTER TABLE threads RENAME CONSTRAINT chk_activities_color_hex         TO chk_threads_color_hex;
ALTER TABLE threads RENAME CONSTRAINT chk_activities_timeout_positive  TO chk_threads_timeout_positive;

ALTER TABLE thread_images RENAME CONSTRAINT activity_images_pkey             TO thread_images_pkey;
ALTER TABLE thread_images RENAME CONSTRAINT activity_images_activity_id_fkey TO thread_images_thread_id_fkey;

ALTER TABLE commitments RENAME CONSTRAINT activity_schedules_pkey              TO commitments_pkey;
ALTER TABLE commitments RENAME CONSTRAINT activity_schedules_activity_id_fkey  TO commitments_thread_id_fkey;
ALTER TABLE commitments RENAME CONSTRAINT uq_activity_schedules_activity_id_id TO uq_commitments_thread_id_id;

ALTER TABLE commitment_histories RENAME CONSTRAINT activity_schedule_histories_pkey             TO commitment_histories_pkey;
ALTER TABLE commitment_histories RENAME CONSTRAINT activity_schedule_histories_activity_id_fkey TO commitment_histories_thread_id_fkey;
ALTER TABLE commitment_histories RENAME CONSTRAINT activity_schedule_histories_schedule_id_fkey TO commitment_histories_commitment_id_fkey;

ALTER TABLE moments RENAME CONSTRAINT activity_captures_pkey                 TO moments_pkey;
ALTER TABLE moments RENAME CONSTRAINT activity_captures_activity_id_fkey     TO moments_thread_id_fkey;
ALTER TABLE moments RENAME CONSTRAINT chk_activity_captures_color_hex        TO chk_moments_color_hex;
ALTER TABLE moments RENAME CONSTRAINT fk_activity_captures_schedule_activity TO fk_moments_commitment_thread;

ALTER TABLE moment_images RENAME CONSTRAINT activity_capture_images_pkey                     TO moment_images_pkey;
ALTER TABLE moment_images RENAME CONSTRAINT activity_capture_images_activity_capture_id_fkey TO moment_images_moment_id_fkey;

-- ---------------------------------------------------------
-- Indexes
-- ---------------------------------------------------------
ALTER INDEX idx_activities_user_id                       RENAME TO idx_threads_user_id;
ALTER INDEX idx_activity_images_activity_id              RENAME TO idx_thread_images_thread_id;
ALTER INDEX idx_activity_schedules_activity_id           RENAME TO idx_commitments_thread_id;
ALTER INDEX idx_activity_schedule_histories_activity_id  RENAME TO idx_commitment_histories_thread_id;
ALTER INDEX idx_activity_schedule_histories_schedule_id  RENAME TO idx_commitment_histories_commitment_id;
ALTER INDEX idx_activity_captures_schedule_id            RENAME TO idx_moments_commitment_id;
ALTER INDEX idx_activity_captures_occurred_at            RENAME TO idx_moments_occurred_at;
ALTER INDEX uq_activity_captures_activity_scheduled_at   RENAME TO uq_moments_thread_id_due_at;
ALTER INDEX idx_activity_capture_images_activity_capture_id RENAME TO idx_moment_images_moment_id;
