CREATE TABLE fighters(
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255),
    height_feet INTEGER,
    height_inches INTEGER,
    weight_lbs INTEGER,
    weight_class VARCHAR(50),
    reach INTEGER,
    leg_reach INTEGER,
    last_fight VARCHAR(20),
    country VARCHAR(100),
    record_wins INTEGER DEFAULT 0,
    record_losses INTEGER DEFAULT 0,
    record_draws INTEGER DEFAULT 0,
    record_nc INTEGER DEFAULT 0,
    fighter_image_url TEXT,

    is_active BOOLEAN,
    created_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT unique_fighter_name UNIQUE(name),
    CONSTRAINT valid_height_feet CHECK (height_feet >=4 AND height_feet <=7)
    CONSTRAINT valid_height_inches CHECK (height_inches >=0 AND height_inches <=11)
);

CREATE INDEX idx_fighter_name ON fighters(name);
CREATE INDEX idx_fighter_weight ON fighters(weight_class);
CREATE INDEX idx_fighter_activity ON fighters(is_active);


CREATE TABLE cards(
    id SERIAL PRIMARY KEY,
    card_name VARCHAR(255) NOT NULL,
    card_number VARCHAR(50),
    location VARCHAR(255),
    event_date DATE,
    created_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_card_name ON cards(card_name);

CREATE TABLE matches(
    id SERIAL PRIMARY KEY,
    card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
    fighter1_id INTEGER NOT NULL REFERENCES fighters(id) ON DELETE RESTRICT,
    fighter2_id INTEGER NOT NULL REFERENCES fighters(id) ON DELETE RESTRICT,
    fight_order INTEGER,
    weight_class VARCHAR(50),
    is_title_fight BOOLEAN DEFAULT false,
    rounds INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT different_fighters CHECK (fighter1_id != fighter2_id)
);

CREATE INDEX idx_matches ON matches(id);
CREATE INDEX idx_matches_fighters ON matches(fighter1_id,fighter2_id);

CREATE TABLE scraper_metadata (
    id SERIAL PRIMARY KEY,
    source_url TEXT NOT NULL,
    last_scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    fighters_scraped INTEGER DEFAULT 0,
    scrape_status VARCHAR(50),
    error_message TEXT,
    CONSTRAINT unique_source_url UNIQUE(source_url)
);

