CREATE SCHEMA main_shard;
CREATE SEQUENCE main_shard.global_id_sequence;

CREATE OR REPLACE FUNCTION main_shard.id_generator(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    -- the id of this DB shard, must be set for each
    -- schema shard you have - you could pass this as a parameter too
    shard_id int := 1;
BEGIN
    SELECT nextval('main_shard.global_id_sequence') % 1024 INTO seq_id;

    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id << 10);
    result := result | (seq_id);
END;
$$ LANGUAGE PLPGSQL;

CREATE TABLE main_shard.users(
    user_id BIGINT NOT NULL DEFAULT main_shard.id_generator(), 
    firebase_uid VARCHAR(32) NOT NULL UNIQUE,
    name VARCHAR(32) NOT NULL DEFAULT '',
    is_customer BOOLEAN NOT NULL DEFAULT false);

CREATE TABLE main_shard.post_tag(
    post_id BIGINT NOT NULL, 
    tag_id BIGINT NOT NULL);

CREATE TABLE main_shard.likes(
    user_id BIGINT NOT NULL, 
    post_id BIGINT NOT NULL);