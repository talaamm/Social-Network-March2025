CREATE TABLE IF NOT EXISTS group_posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    group_id  INTEGER NOT NULL,
    member_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    username TEXT NOT NULL,
    image TEXT DEFAULT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    foreign KEY(group_id) references groups(id) ON DELETE CASCADE,
    FOREIGN KEY (member_id, group_id) REFERENCES group_members(id, group_id) ON DELETE CASCADE
    -- foreign KEY(member_id) references group_members(id) ON DELETE CASCADE
);