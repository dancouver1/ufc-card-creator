-- Create rankings table: stores both "media" and "meta" (AI) UFC rankings per division
CREATE TABLE rankings (
    id SERIAL PRIMARY KEY,
    rank_type VARCHAR(10) NOT NULL,
    division VARCHAR(50) NOT NULL,
    position INTEGER NOT NULL,
    fighter_name VARCHAR(255) NOT NULL,
    fighter_slug VARCHAR(255),
    is_champion BOOLEAN DEFAULT false,
    movement INTEGER DEFAULT 0,
    champion_image_url TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_rank_type CHECK (rank_type IN ('media', 'meta')),
    -- Rankings can tie (two fighters sharing a position), so fighter_name is
    -- part of the uniqueness key alongside the slot.
    CONSTRAINT unique_ranking_slot UNIQUE(rank_type, division, position, fighter_name)
);

CREATE INDEX idx_rankings_type_division ON rankings(rank_type, division);
