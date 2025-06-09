-- CREATE TABLE IF NOT EXISTS group_members (
--     id INTEGER NOT NULL PRIMARY KEY,
--     group_id INTEGER NOT NULL,
--     username TEXT NOT NULL,
--     status TEXT CHECK(status IN ('pending', 'approved' , 'rejected')) DEFAULT 'pending',
--     FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
--     FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
-- );

CREATE TABLE IF NOT EXISTS group_members (
    id INTEGER NOT NULL,  -- User ID (same as users.id)
    group_id INTEGER NOT NULL,  -- Group ID (from groups table)
    username TEXT NOT NULL,
    status TEXT CHECK(status IN ('pending', 'approved', 'rejected')) DEFAULT 'pending',
    PRIMARY KEY (id, group_id),  -- ✅ Users can join multiple groups, no duplicates
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);
