CREATE TABLE post_tag(post_id INT, tag_id INT);
CREATE TABLE likes(user_id INT, post_id INT);
CREATE TABLE users(user_id SERIAL, firebase_uid VARCHAR(32));

--add testing data
INSERT INTO likes(user_id, post_id) VALUES (1, 1), (1, 3), (1, 5);
INSERT INTO post_tag(post_id, tag_id) VALUES (1, 1), (1, 2), (2, 1), (2, 2), (3, 1), (3, 3), (4, 1), (4, 5), (5, 1), (5, 4), (6, 6), (6, 2), (7, 7), (7, 8);