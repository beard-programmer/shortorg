CREATE TABLE encoded_urls (
                              token_identifier BIGINT PRIMARY KEY,
                              url VARCHAR(255) NOT NULL,
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
