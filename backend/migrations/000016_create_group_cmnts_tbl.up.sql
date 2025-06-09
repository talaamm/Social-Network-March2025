CREATE TABLE IF NOT EXISTS group_comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    member_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    g_post_id INTEGER NOT NULL,
    content TEXT,
    image TEXT,
    username TEXT,
    FOREIGN KEY (member_id, group_id) REFERENCES group_members(id, group_id) ON DELETE CASCADE,
    FOREIGN KEY (g_post_id) REFERENCES group_posts(id) ON DELETE CASCADE
);