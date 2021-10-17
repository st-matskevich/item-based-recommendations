CREATE TABLE post_tag(post_id BIGINT NOT NULL, tag_id BIGINT NOT NULL);
CREATE TABLE likes(user_id BIGINT NOT NULL, post_id BIGINT NOT NULL);
CREATE TABLE users(user_id BIGINT NOT NULL, firebase_uid VARCHAR(32) NOT NULL);

--add testing data
INSERT INTO likes(user_id, post_id) VALUES (1, 1), (1, 3), (1, 5);
INSERT INTO post_tag(post_id, tag_id) VALUES (1, 1), (1, 2), (2, 1), (2, 2), (3, 1), (3, 3), (4, 1), (4, 5), (5, 1), (5, 4), (6, 6), (6, 2), (7, 7), (7, 8);