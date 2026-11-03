CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL DEFAULT '',
    last_name VARCHAR(255) NOT NULL DEFAULT '',
    password VARCHAR(255) NOT NULL,
    user_active INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Demo user: admin@example.com / verysecret
INSERT INTO users (email, first_name, last_name, password, user_active)
VALUES (
    'admin@example.com',
    'Admin',
    'User',
    '$2a$12$O8lffwP3SXVHg8Lkt3OboeQSDg4hdezHsHpMc.cPae7mbPeWMh4uO',
    1
)
ON CONFLICT (email) DO NOTHING;
