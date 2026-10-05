DROP TABLE IF EXISTS roles;
CREATE TABLE IF NOT EXISTS roles (
    name VARCHAR(50) PRIMARY KEY,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO roles (name, description) VALUES
('admin', 'Administrator sistem dengan wewenang penuh'),
('user', 'Pengguna reguler layanan obrolan')
ON CONFLICT (name) DO NOTHING;