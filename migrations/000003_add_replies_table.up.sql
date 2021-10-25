CREATE TABLE replies(
    reply_id BIGINT NOT NULL DEFAULT id_generator(),
    task_id BIGINT NOT NULL,
    text VARCHAR(512) NOT NULL,
    creator_id BIGINT NOT NULL,
    hidden BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (task_id, creator_id));