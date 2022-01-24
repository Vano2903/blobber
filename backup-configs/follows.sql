CREATE TABLE IF NOT EXIST follows (
    ID INT auto_increment NOT NULL,
    ID_user_follower INT NOT NULL,
    ID_user_followed INT NOT NULL,
    PRIMARY KEY (ID)
);
-- ENGINE=InnoDB
-- DEFAULT CHARSET=utf8mb4
-- COLLATE=utf8mb4_unicode_ci