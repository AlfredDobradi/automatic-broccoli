CREATE TABLE IF NOT EXISTS messages (
	time TIMESTAMPTZ NOT NULL,
    recipient VARCHAR(30) NOT NULL DEFAULT 'system',
    sender VARCHAR(30) NOT NULL DEFAULT 'global',
    type VARCHAR(30) NOT NULL DEFAULT 'chat',
    body TEXT NOT NULL DEFAULT ''
);

SELECT create_hypertable('messages', 'time')