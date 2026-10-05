DROP TABLE IF EXISTS permissions;
CREATE TABLE IF NOT EXISTS permissions (
    name VARCHAR(100) PRIMARY KEY,
    description TEXT
);

DROP TABLE IF EXISTS role_permissions;
CREATE TABLE IF NOT EXISTS role_permissions (
    role_name VARCHAR(50) NOT NULL REFERENCES roles(name) ON DELETE CASCADE,
    permission_name VARCHAR(100) NOT NULL REFERENCES permissions(name) ON DELETE CASCADE,
    PRIMARY KEY (role_name, permission_name)
);

INSERT INTO permissions (name, description) VALUES
('user:list', 'Melihat daftar seluruh pengguna'),
('user:create', 'Membuat akun pengguna baru'),
('user:read:any', 'Melihat detail akun pengguna mana pun'),
('user:update:any', 'Mengubah data akun pengguna mana pun'),
('user:delete', 'Menonaktifkan akun pengguna mana pun'),
('role:assign', 'Mengubah role pengguna'),
('chat:room:access', 'Mengakses ruang obrolan'),
('chat:message:send', 'Mengirim pesan obrolan')
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_name, permission_name) VALUES
('admin', 'user:list'),
('admin', 'user:create'),
('admin', 'user:read:any'),
('admin', 'user:update:any'),
('admin', 'user:delete'),
('admin', 'role:assign'),
('admin', 'chat:room:access'),
('admin', 'chat:message:send'),
('user', 'chat:room:access'),
('user', 'chat:message:send')
ON CONFLICT DO NOTHING;