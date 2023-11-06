CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT,
    password_hash TEXT,
    verification_code TEXT,
    is_verified BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE links (
    link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    original_url TEXT,
    redirect_path TEXT,
    email_subject TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    number_of_times_clicked INTEGER DEFAULT 0
);