CREATE TABLE notifications(
    notification_id BIGINT NOT NULL DEFAULT id_generator(),
    user_id BIGINT NOT NULL,
    trigger_id BIGINT NOT NULL,
    type INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now());