-- Insert some test fighters
INSERT INTO fighters (name, nickname, height_feet, height_inches, weight_lbs, reach_cm, weight_class, stance, wins, losses, draws, is_active)
VALUES 
    ('Conor McGregor', 'The Notorious', 5, 9, 155, 188, 'Lightweight', 'Southpaw', 22, 6, 0, true),
    ('Khabib Nurmagomedov', 'The Eagle', 5, 10, 155, 178, 'Lightweight', 'Orthodox', 29, 0, 0, true),
    ('Israel Adesanya', 'The Last Stylebender', 6, 4, 185, 203, 'Middleweight', 'Switch', 24, 3, 0, true),
    ('Alex Pereira', 'Poatan', 6, 4, 185, 201, 'Middleweight', 'Orthodox', 9, 2, 0, true),
    ('Jon Jones', 'Bones', 6, 4, 240, 215, 'Heavyweight', 'Orthodox', 27, 1, 0, true),
    ('Francis Ngannou', 'The Predator', 6, 4, 257, 211, 'Heavyweight', 'Orthodox', 17, 3, 0, true),
    ('Alexander Volkanovski', 'The Great', 5, 6, 145, 183, 'Featherweight', 'Orthodox', 26, 3, 0, true),
    ('Max Holloway', 'Blessed', 5, 11, 145, 175, 'Featherweight', 'Orthodox', 25, 7, 0, true),
    ('Leon Edwards', 'Rocky', 6, 0, 170, 188, 'Welterweight', 'Orthodox', 22, 3, 0, true),
    ('Kamaru Usman', 'The Nigerian Nightmare', 6, 0, 170, 193, 'Welterweight', 'Orthodox', 20, 4, 0, true);

-- Add some fight history for testing the "Last 5" feature
INSERT INTO fight_history (fighter_id, opponent_name, result, fight_order)
VALUES
    -- Conor McGregor (ID will be 1)
    (1, 'Dustin Poirier', 'loss', 5),
    (1, 'Dustin Poirier', 'loss', 4),
    (1, 'Donald Cerrone', 'win', 3),
    (1, 'Khabib Nurmagomedov', 'loss', 2),
    (1, 'Eddie Alvarez', 'win', 1),
    
    -- Khabib Nurmagomedov (ID will be 2)
    (2, 'Justin Gaethje', 'win', 5),
    (2, 'Dustin Poirier', 'win', 4),
    (2, 'Conor McGregor', 'win', 3),
    (2, 'Al Iaquinta', 'win', 2),
    (2, 'Edson Barboza', 'win', 1),
    
    -- Israel Adesanya (ID will be 3)
    (3, 'Alex Pereira', 'loss', 5),
    (3, 'Jared Cannonier', 'win', 4),
    (3, 'Robert Whittaker', 'win', 3),
    (3, 'Marvin Vettori', 'win', 2),
    (3, 'Paulo Costa', 'win', 1);