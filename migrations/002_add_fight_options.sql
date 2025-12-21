-- Add columns for fight options
ALTER TABLE matches 
ADD COLUMN is_main_event BOOLEAN DEFAULT false,
ADD COLUMN is_co_main_event BOOLEAN DEFAULT false,
ADD COLUMN title_type VARCHAR(50), -- 'Undisputed', 'Interim', 'BMF'
ADD COLUMN card_part VARCHAR(50) DEFAULT 'Main Card'; -- 'Main Card', 'Prelims'
