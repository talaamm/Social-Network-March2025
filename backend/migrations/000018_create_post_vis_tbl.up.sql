CREATE TABLE IF NOT EXISTS posts_visibility  (
post_creator INTEGER NOT NULL,
user_id INTEGER NOT NULL,
FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
FOREIGN KEY (post_creator) REFERENCES users(id) ON DELETE CASCADE,
PRIMARY KEY (post_creator, user_id) --composit primary key
);
