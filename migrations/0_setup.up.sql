CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    password_hash TEXT,
    verification_code UUID,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX index_users_email_verification_code
ON users (email, verification_code);

CREATE TABLE links (
    link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    original_url TEXT,
    redirect_path TEXT UNIQUE,
    email_subject TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX index_links_redirect_path
ON links (redirect_path);

CREATE INDEX index_links_user_id
ON links (user_id);

CREATE TABLE clicks (
    click_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID REFERENCES links(link_id),
    clicked_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX clicks_index
ON clicks (link_id);