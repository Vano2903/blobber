CREATE TABLE IF NOT EXIST likes (
    ID INT auto_increment NOT NULL,
    ID_user INT NOT NULL,
    ID_blob INT NOT NULL,
    PRIMARY KEY (ID)
);
-- ENGINE=InnoDB
-- DEFAULT CHARSET=utf8mb4
-- COLLATE=utf8mb4_unicode_ci;