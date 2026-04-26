SET search_path TO public;

GRANT ALL ON ALL TABLES IN SCHEMA public TO "user";
GRANT USAGE ON SCHEMA public TO "user";
ALTER USER "user" SET search_path TO postgres, public;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT ALL ON TABLES TO "user";

GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO "user";

ALTER DEFAULT PRIVILEGES IN SCHEMA public
GRANT USAGE, SELECT ON SEQUENCES TO "user";


CREATE TABLE posts  (
    post_id SERIAL PRIMARY KEY,
    title VARCHAR(255),
    text_content TEXT,
    OP_id INTEGER NULL,
    created_at TIMESTAMP NOT NULL,
    last_updated_at TIMESTAMP NOT NULL
);

CREATE TABLE anons (
    anon_id SERIAL PRIMARY KEY,
    name VARCHAR(255),
    post_id INTEGER NOT NULL,
    avatar VARCHAR(255)
);

ALTER TABLE posts
    ADD CONSTRAINT fk_posts_op_id
    FOREIGN KEY (OP_id) REFERENCES anons (anon_id);

ALTER TABLE anons
    ADD CONSTRAINT fk_anons_post_id
    FOREIGN KEY (post_id) REFERENCES posts (post_id);

CREATE INDEX idx_posts_created_at ON posts (created_at DESC);
CREATE INDEX idx_posts_last_updated_at ON posts (last_updated_at DESC);
CREATE INDEX idx_anons_post_id_avatar ON anons (post_id, avatar);

CREATE TABLE public.comments (
    comment_id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts (post_id),
    addressed_to INTEGER NOT NULL,
    text_content TEXT NOT NULL,
    anon_id INTEGER NOT NULL REFERENCES anons (anon_id),
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_comments_post_id_addressed_to_created_at ON comments (post_id, addressed_to, created_at);

CREATE TABLE public.attachments (
    attachment_id SERIAL PRIMARY KEY,
    post_id INTEGER NOT NULL REFERENCES posts (post_id),
    comment_id INTEGER REFERENCES comments (comment_id),
    file_key VARCHAR(255),
    original_name VARCHAR(255),
    content_type VARCHAR(255),
    size INTEGER,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_attachments_post_id_comment_id ON attachments (post_id, comment_id);
CREATE INDEX idx_attachments_comment_id ON attachments (comment_id);
CREATE UNIQUE INDEX idx_attachments_file_key ON attachments (file_key);

CREATE TABLE public.sessions (
    session_id SERIAL PRIMARY KEY,
    session_history JSON,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_expires_at ON sessions (expires_at);


