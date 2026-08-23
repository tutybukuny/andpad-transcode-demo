BEGIN;

CREATE TABLE IF NOT EXISTS tdapp.transcode_request
(
    id                    bigserial PRIMARY KEY NOT NULL,
    video_url             TEXT                  NOT NULL,
    output_url            TEXT                  NOT NULL,
    master_file_url       TEXT                  NOT NULL,
    status                VARCHAR(20)           NOT NULL DEFAULT 'todo',
    failed_reason         TEXT                  NOT NULL DEFAULT '',
    started_transcode_at  TIMESTAMPTZ,
    finished_transcode_at TIMESTAMPTZ,
    last_processing_at    TIMESTAMPTZ,
    retried_times         INT                   NOT NULL DEFAULT 0,
    created_at            TIMESTAMPTZ                    DEFAULT NOW() NOT NULL,
    updated_at            TIMESTAMPTZ                    DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transcode_request_status ON tdapp.transcode_request (status);
CREATE INDEX IF NOT EXISTS idx_transcode_request_created_at ON tdapp.transcode_request (updated_at);
CREATE INDEX IF NOT EXISTS idx_transcode_request_last_processing_retried ON tdapp.transcode_request (last_processing_at, retried_times);

COMMIT;