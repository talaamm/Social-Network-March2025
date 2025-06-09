CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    nickname TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    age INTEGER DEFAULT 0,
    gender TEXT DEFAULT "Not Specified",
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    date_of_birth TEXT Not Null,
    is_private BOOLEAN DEFAULT FALSE
);
