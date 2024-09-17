CREATE TABLE encoded_urls (
                              token_identifier BIGINT PRIMARY KEY,
                              token VARCHAR(7) NOT NULL UNIQUE,
                              url VARCHAR(255) NOT NULL,
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

