CREATE TABLE notifications(
    notification_id BIGINT PRIMARY KEY NOT NULL DEFAULT id_generator(),
    user_id BIGINT NOT NULL,
    trigger_id BIGINT NOT NULL,
    type INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
            REFERENCES users(user_id)
                ON DELETE CASCADE);