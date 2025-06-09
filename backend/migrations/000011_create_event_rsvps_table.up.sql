CREATE TABLE IF NOT EXISTS group_event_attendees (
    member_id INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    status TEXT CHECK(status IN ('going', 'not going')),
    PRIMARY KEY (event_id, member_id),
    FOREIGN KEY (group_id) REFERENCES groups(id),
    FOREIGN KEY (member_id, group_id) REFERENCES group_members(id, group_id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE
);