CREATE TABLE topic_aliases (
    id INTEGER PRIMARY KEY AUTO_INCREMENT,
    topic_id INTEGER NOT NULL,
    alias_name VARCHAR(75) NOT NULL UNIQUE,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CHECK (length(alias_name) > 0)
);
