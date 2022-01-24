CREATE TABLE IF NOT EXIST users (
    ID INT auto_increment NOT NULL,
    username VARCHAR(20) NOT NULL,
    password CHAR(64) NOT NULL,
    description TEXT,
    PRIMARY KEY (ID)
);
-- ENGINE=InnoDB
-- DEFAULT CHARSET=utf8mb4
-- COLLATE=utf8mb4_unicode_ci;