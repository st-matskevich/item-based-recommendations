DROP INDEX tag_text_index;
ALTER TABLE task_tag DROP CONSTRAINT task_tag_unique;
ALTER TABLE likes DROP CONSTRAINT likes_user_task;