CREATE TABLE IF NOT EXISTS app_conofig (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
)

-- seed default value
INSERT INTO app_conofig (key, value)
VALUES ('max_load', '5')
ON CONFLICT (key) DO NOTHING;