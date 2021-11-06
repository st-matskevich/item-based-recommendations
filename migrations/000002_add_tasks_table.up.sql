CREATE TABLE tasks(
    task_id BIGINT PRIMARY KEY NOT NULL DEFAULT id_generator(),
    name VARCHAR(64) NOT NULL,
    description VARCHAR(512) NOT NULL,
    customer_id BIGINT NOT NULL,
    doer_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT fk_customer
        FOREIGN KEY(customer_id) 
            REFERENCES users(user_id)
                ON DELETE CASCADE,
    CONSTRAINT fk_doer
        FOREIGN KEY(doer_id) 
            REFERENCES users(user_id)
                ON DELETE SET NULL);