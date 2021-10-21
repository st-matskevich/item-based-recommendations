CREATE TABLE posts(
    post_id BIGINT NOT NULL DEFAULT id_generator(),
    name VARCHAR(64) NOT NULL,
    description VARCHAR(512) NOT NULL,
    customer_id BIGINT NOT NULL,
    doer_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT now());