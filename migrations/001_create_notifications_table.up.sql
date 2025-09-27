CREATE TABLE IF NOT EXISTS notifications (
    uuid TEXT PRIMARY KEY,
    text_message TEXT,
    publish_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status_publish TEXT NOT NULL,
    tg_chatid BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);