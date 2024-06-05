CREATE SCHEMA IF NOT EXISTS auth;
CREATE TABLE IF NOT EXISTS auth.users(
                      id BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY ,
                      username VARCHAR(255) ,
                      email VARCHAR(255) NOT NULL ,
                      password_hash bytea NOT NULL ,
                      user_role VARCHAR(50),
                      activated bool NOT NULL
);