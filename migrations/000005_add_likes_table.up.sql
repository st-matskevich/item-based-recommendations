CREATE TABLE likes(
    like_id BIGINT PRIMARY KEY NOT NULL DEFAULT id_generator(),
    user_id BIGINT NOT NULL, 
    task_id BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
            REFERENCES users(user_id)
                ON DELETE CASCADE,
    CONSTRAINT fk_task
        FOREIGN KEY(task_id) 
            REFERENCES tasks(task_id)
                ON DELETE CASCADE);