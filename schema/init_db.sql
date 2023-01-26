-- install extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- create table
CREATE TABLE secrets
(
    id              uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    title           varchar(255) NOT NULL,
    message         varchar(255) NOT NULL,
    passphrase      bytea        NOT NULL,
    iv              bytea        NOT NULL,
    expiration      timestamp with time zone,
    reads_remaining integer      NOT NULL,
    owner           varchar(255)
);


-- This script creates a table called secrets with the following fields:
-- * id: a serial primary key
-- * name: the name of the secret (not null)
-- * username: the username associated with the secret (not null)
-- * password: the encrypted password of the secret (not null), i.e. the secret itself (TODO: rename)
-- * iv: the initialization vector used for the encryption of the password (not null)
-- * expiration: the expiration date of the secret
-- * reads_remaining: the number of times the secret can be read (not null)
-- * owner: the owner of the secret

-- populate DB with data
-- INSERT INTO secrets (name, username, password, iv, expiration, reads_remaining, owner)
-- VALUES ('')