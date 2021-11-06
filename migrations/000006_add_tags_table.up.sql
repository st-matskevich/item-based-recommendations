CREATE TABLE tags(
    tag_id BIGINT PRIMARY KEY NOT NULL DEFAULT id_generator(),
    text VARCHAR(32) NOT NULL DEFAULT '');

CREATE TABLE task_tag(
    task_id BIGINT NOT NULL, 
    tag_id BIGINT NOT NULL,
    CONSTRAINT fk_task
        FOREIGN KEY(task_id) 
            REFERENCES tasks(task_id)
                ON DELETE CASCADE,
    CONSTRAINT fk_tag
        FOREIGN KEY(tag_id) 
            REFERENCES tags(tag_id)
                ON DELETE CASCADE);