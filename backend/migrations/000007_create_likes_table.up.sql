-- CREATE TABLE IF NOT EXISTS likes (
--     id INTEGER PRIMARY KEY AUTOINCREMENT,
--     user_id INTEGER NOT NULL,
--     post_id INTEGER DEFAULT NULL,
--     comment_id INTEGER DEFAULT NULL, --not needed at all
--     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
--     FOREIGN KEY (user_id) REFERENCES users(id),
--     FOREIGN KEY (post_id) REFERENCES posts(id),
--     FOREIGN KEY (comment_id) REFERENCES comments(id),
--     UNIQUE(user_id, post_id, comment_id) -- Prevents duplicate likes
-- );

CREATE TABLE IF NOT EXISTS likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER DEFAULT NULL,
    user_id INTEGER  NOT NULL,
    is_like BOOLEAN DEFAULT NULL, -- true for like, false for dislike
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);