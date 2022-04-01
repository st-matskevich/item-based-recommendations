CREATE UNIQUE INDEX tag_text_index ON tags(text);
ALTER TABLE task_tag ADD CONSTRAINT task_tag_unique UNIQUE (task_id, tag_id);
ALTER TABLE likes ADD CONSTRAINT likes_user_task UNIQUE (user_id, task_id);