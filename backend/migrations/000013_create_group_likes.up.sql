
CREATE TABLE IF NOT EXISTS group_likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER DEFAULT NULL,
    member_id INTEGER  NOT NULL,
    is_like BOOLEAN DEFAULT NULL, -- true for like, false for dislike
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);