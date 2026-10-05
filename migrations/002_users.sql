DROP TABLE IF EXISTS users;
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user' REFERENCES roles(name) ON UPDATE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at_id ON users(created_at DESC, id DESC);

INSERT INTO users (username, email, password, role, is_active)
VALUES (
    'admin',
    'admin@gotalk.com',
    '$2a$12$D23M1hA78gHhT5fG7N1M6uebK1YgB6uEcv48T.g1qS1sE1YqOqRia',
    'admin',
    true
) ON CONFLICT (email) DO NOTHING;