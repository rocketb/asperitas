INSERT INTO users (user_id, name, roles, password_hash, date_created) VALUES
    ('5cf37266-3473-4006-984f-9325122678b7', 'admin', '{ADMIN,USER}', '$2a$10$86a8En9ddIQI7t2acorxJenWP7SShjIGXUXCYZtpz.iDNmSGdBCcq', '2023-01-21 00:00:00'),
    ('5cf37266-3473-4006-984f-9325122678b9', 'user', '{USER}', '$2a$10$86a8En9ddIQI7t2acorxJenWP7SShjIGXUXCYZtpz.iDNmSGdBCcq', '2023-01-21 00:00:00')
    ON CONFLICT DO NOTHING;

INSERT INTO posts (post_id, type, title, category, body, views, date_created, user_id) VALUES
    ('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', 'text', 'Text post title', 'books', 'Text.', 1, '2023-01-22 00:00:00', '5cf37266-3473-4006-984f-9325122678b9'),
    ('72f8b983-3eb4-48db-9ed0-e45cc6bd716b', 'url', 'Url post title', 'music', 'https://music.com', 1, '2023-01-24 00:00:00', '5cf37266-3473-4006-984f-9325122678b9')
    ON CONFLICT DO NOTHING;

INSERT INTO comments (comment_id, post_id, user_id, body, date_created) VALUES
    ('98b6d4b8-f04b-4c79-8c2e-a0aef46854b7', 'a2b0639f-2cc6-44b8-b97b-15d69dbb511e', '5cf37266-3473-4006-984f-9325122678b9', 'Awesome post!', '2023-01-23 00:00:00')
    ON CONFLICT DO NOTHING;

INSERT INTO votes (post_id, user_id, vote) VALUES
    ('a2b0639f-2cc6-44b8-b97b-15d69dbb511e', '5cf37266-3473-4006-984f-9325122678b9', 1)
    ON CONFLICT DO NOTHING;
