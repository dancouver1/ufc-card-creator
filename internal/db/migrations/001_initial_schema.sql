-- Create fighters table
CREATE TABLE fighters (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255),
    height_feet INTEGER,
    height_inches INTEGER,
    weight_lbs INTEGER,
    reach_cm INTEGER,
    leg_reach_cm INTEGER,
    weight_class VARCHAR(50) NOT NULL,
    stance VARCHAR(20),
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    draws INTEGER DEFAULT 0,
    date_of_birth DATE,
    nationality VARCHAR(100),
    fighter_image_url TEXT,
    is_active BOOLEAN DEFAULT true,
    last_fight_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_fighter_name UNIQUE(name),
    CONSTRAINT valid_height_feet CHECK (height_feet >= 4 AND height_feet <= 7),
    CONSTRAINT valid_height_inches CHECK (height_inches >= 0 AND height_inches <= 11)
);

CREATE INDEX idx_fighters_weight_class ON fighters(weight_class);
CREATE INDEX idx_fighters_name ON fighters(name);
CREATE INDEX idx_fighters_active ON fighters(is_active);

-- Create cards table
CREATE TABLE cards (
    id SERIAL PRIMARY KEY,
    card_name VARCHAR(255) NOT NULL,
    card_number VARCHAR(50),
    location VARCHAR(255),
    event_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cards_name ON cards(card_name);

-- Create matches table
CREATE TABLE matches (
    id SERIAL PRIMARY KEY,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    fighter1_id INTEGER NOT NULL REFERENCES fighters(id) ON DELETE RESTRICT,
    fighter2_id INTEGER NOT NULL REFERENCES fighters(id) ON DELETE RESTRICT,
    fight_order INTEGER,
    weight_class VARCHAR(50),
    is_title_fight BOOLEAN DEFAULT false,
    rounds INTEGER DEFAULT 3,
    prediction INTEGER REFERENCES fighters(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT different_fighters CHECK (fighter1_id != fighter2_id),
    CONSTRAINT valid_prediction CHECK (prediction IS NULL OR prediction = fighter1_id OR prediction = fighter2_id)
);

CREATE INDEX idx_matches_card ON matches(card_id);
CREATE INDEX idx_matches_fighters ON matches(fighter1_id, fighter2_id);

-- Create fight_history table
CREATE TABLE fight_history (
    id SERIAL PRIMARY KEY,
    fighter_id INTEGER NOT NULL REFERENCES fighters(id) ON DELETE CASCADE,
    opponent_name VARCHAR(255) NOT NULL,
    result VARCHAR(10) NOT NULL,
    method VARCHAR(100),
    round INTEGER,
    fight_date DATE,
    event_name VARCHAR(255),
    fight_order INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_result CHECK (result IN ('win', 'loss', 'draw', 'nc'))
);

CREATE INDEX idx_fight_history_fighter ON fight_history(fighter_id);
CREATE INDEX idx_fight_history_date ON fight_history(fight_date DESC);

-- Create scraper metadata table
CREATE TABLE scraper_metadata (
    id SERIAL PRIMARY KEY,
    source_url TEXT NOT NULL,
    last_scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    fighters_scraped INTEGER DEFAULT 0,
    scrape_status VARCHAR(50),
    error_message TEXT,
    CONSTRAINT unique_source_url UNIQUE(source_url)
);