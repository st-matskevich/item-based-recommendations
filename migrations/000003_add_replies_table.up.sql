CREATE TABLE replies(
    reply_id BIGINT PRIMARY KEY NOT NULL DEFAULT id_generator(),
    task_id BIGINT NOT NULL,
    text VARCHAR(512) NOT NULL,
    creator_id BIGINT NOT NULL,
    hidden BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    UNIQUE (task_id, creator_id),
    CONSTRAINT fk_creator
        FOREIGN KEY(creator_id) 
            REFERENCES users(user_id)
                ON DELETE CASCADE,
    CONSTRAINT fk_task
        FOREIGN KEY(task_id) 
            REFERENCES tasks(task_id)
                ON DELETE CASCADE);