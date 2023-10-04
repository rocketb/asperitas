-- Version: 1.01
-- Description: Create users table
CREATE TABLE users (
    user_id        UUID        NOT NULL,
    name           TEXT UNIQUE NOT NULL,
    roles          TEXT[]      NOT NULL,
    password_hash  TEXT        NOT NULL,
    date_created   TIMESTAMP   NOT NULL,

    PRIMARY KEY (user_id)
);

-- Version: 1.02
-- Description: Create posts table
CREATE TABLE posts (
    post_id        UUID      NOT NULL,
    type           TEXT      NOT NULL,
    title          TEXT      NOT NULL,
    category       TEXT      NOT NULL,
    body           TEXT      NOT NULL,
    views          INT       NOT NULL,
    date_created   TIMESTAMP NOT NULL,
    user_id        UUID      NOT NULL,

    PRIMARY KEY (post_id),
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Version: 1.03
-- Description: Create comments table
CREATE TABLE comments (
    comment_id     UUID      NOT NULL,
    post_id        UUID      NOT NULL,
    user_id        UUID      NOT NULL,
    body           TEXT      NOT NULL,
    date_created   TIMESTAMP NOT NULL,

    PRIMARY KEY (comment_id),
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE
);

-- Version: 1.04
-- Description: Create votes table
CREATE TABLE votes (
    post_id        UUID NOT NULL,
    user_id        UUID NOT NULL,
    vote           INT  NOT NULL,

    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
)
