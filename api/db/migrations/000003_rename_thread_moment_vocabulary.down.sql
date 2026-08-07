-- Exact inverse of 000003_rename_thread_moment_vocabulary.up.sql.
-- Reverses in the opposite order: indexes, constraints, columns, tables,
-- then the enum.

-- ---------------------------------------------------------
-- Indexes
-- ---------------------------------------------------------
ALTER INDEX idx_moment_images_moment_id           RENAME TO idx_activity_capture_images_activity_capture_id;
ALTER INDEX uq_moments_thread_id_due_at           RENAME TO uq_activity_captures_activity_scheduled_at;
ALTER INDEX idx_moments_occurred_at               RENAME TO idx_activity_captures_occurred_at;
ALTER INDEX idx_moments_commitment_id             RENAME TO idx_activity_captures_schedule_id;
ALTER INDEX idx_commitment_histories_commitment_id RENAME TO idx_activity_schedule_histories_schedule_id;
ALTER INDEX idx_commitment_histories_thread_id    RENAME TO idx_activity_schedule_histories_activity_id;
ALTER INDEX idx_commitments_thread_id             RENAME TO idx_activity_schedules_activity_id;
ALTER INDEX idx_thread_images_thread_id           RENAME TO idx_activity_images_activity_id;
ALTER INDEX idx_threads_user_id                   RENAME TO idx_activities_user_id;

-- ---------------------------------------------------------
-- Constraints
-- ---------------------------------------------------------
ALTER TABLE moment_images RENAME CONSTRAINT moment_images_moment_id_fkey TO activity_capture_images_activity_capture_id_fkey;
ALTER TABLE moment_images RENAME CONSTRAINT moment_images_pkey           TO activity_capture_images_pkey;

ALTER TABLE moments RENAME CONSTRAINT fk_moments_commitment_thread TO fk_activity_captures_schedule_activity;
ALTER TABLE moments RENAME CONSTRAINT chk_moments_color_hex        TO chk_activity_captures_color_hex;
ALTER TABLE moments RENAME CONSTRAINT moments_thread_id_fkey       TO activity_captures_activity_id_fkey;
ALTER TABLE moments RENAME CONSTRAINT moments_pkey                 TO activity_captures_pkey;

ALTER TABLE commitment_histories RENAME CONSTRAINT commitment_histories_commitment_id_fkey TO activity_schedule_histories_schedule_id_fkey;
ALTER TABLE commitment_histories RENAME CONSTRAINT commitment_histories_thread_id_fkey     TO activity_schedule_histories_activity_id_fkey;
ALTER TABLE commitment_histories RENAME CONSTRAINT commitment_histories_pkey               TO activity_schedule_histories_pkey;

ALTER TABLE commitments RENAME CONSTRAINT uq_commitments_thread_id_id TO uq_activity_schedules_activity_id_id;
ALTER TABLE commitments RENAME CONSTRAINT commitments_thread_id_fkey  TO activity_schedules_activity_id_fkey;
ALTER TABLE commitments RENAME CONSTRAINT commitments_pkey            TO activity_schedules_pkey;

ALTER TABLE thread_images RENAME CONSTRAINT thread_images_thread_id_fkey TO activity_images_activity_id_fkey;
ALTER TABLE thread_images RENAME CONSTRAINT thread_images_pkey           TO activity_images_pkey;

ALTER TABLE threads RENAME CONSTRAINT chk_threads_timeout_positive TO chk_activities_timeout_positive;
ALTER TABLE threads RENAME CONSTRAINT chk_threads_color_hex        TO chk_activities_color_hex;
ALTER TABLE threads RENAME CONSTRAINT threads_user_id_fkey         TO activities_user_id_fkey;
ALTER TABLE threads RENAME CONSTRAINT threads_pkey                 TO activities_pkey;

-- ---------------------------------------------------------
-- Columns
-- ---------------------------------------------------------
ALTER TABLE moment_images RENAME COLUMN moment_id TO activity_capture_id;

ALTER TABLE moments RENAME COLUMN captured_at   TO confirmed_at;
ALTER TABLE moments RENAME COLUMN due_at        TO scheduled_at;
ALTER TABLE moments RENAME COLUMN commitment_id TO schedule_id;
ALTER TABLE moments RENAME COLUMN thread_id     TO activity_id;

ALTER TABLE commitment_histories RENAME COLUMN commitment_id TO schedule_id;
ALTER TABLE commitment_histories RENAME COLUMN thread_id     TO activity_id;

ALTER TABLE commitments RENAME COLUMN thread_id TO activity_id;

ALTER TABLE thread_images RENAME COLUMN thread_id TO activity_id;

ALTER TABLE threads RENAME COLUMN has_commitment TO is_fixed_schedule;

-- ---------------------------------------------------------
-- Tables
-- ---------------------------------------------------------
ALTER TABLE moment_images        RENAME TO activity_capture_images;
ALTER TABLE moments              RENAME TO activity_captures;
ALTER TABLE commitment_histories RENAME TO activity_schedule_histories;
ALTER TABLE commitments          RENAME TO activity_schedules;
ALTER TABLE thread_images        RENAME TO activity_images;
ALTER TABLE threads              RENAME TO activities;

-- ---------------------------------------------------------
-- Enum
-- ---------------------------------------------------------
ALTER TYPE moment_status RENAME VALUE 'kept' TO 'captured';
ALTER TYPE moment_status RENAME VALUE 'due'  TO 'awaiting';

ALTER TYPE moment_status RENAME TO activity_capture_status;
