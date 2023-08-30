CREATE TABLE messages
(
    id          BIGINT NOT NULL AUTO_INCREMENT,
    sender_id   BIGINT NOT NULL,
    receiver_id BIGINT NOT NULL,
    message_text TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE
);