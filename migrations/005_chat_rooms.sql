DROP TABLE IF EXISTS chat_rooms;
CREATE TABLE IF NOT EXISTS chat_rooms (
    id BIGSERIAL PRIMARY KEY,
    user1_id INT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    user2_id INT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_canonical_room_users CHECK (user1_id IS NULL OR user2_id IS NULL OR user1_id < user2_id),
    CONSTRAINT uq_canonical_room UNIQUE (user1_id, user2_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_rooms_user1 ON chat_rooms(user1_id);
CREATE INDEX IF NOT EXISTS idx_chat_rooms_user2 ON chat_rooms(user2_id);