CREATE TABLE topic_aliases (
    id SERIAL PRIMARY KEY,
    topic_id SERIAL,
    alias_name VARCHAR(100) NOT NULL UNIQUE,
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CHECK (length(alias_name) > 0)
);
