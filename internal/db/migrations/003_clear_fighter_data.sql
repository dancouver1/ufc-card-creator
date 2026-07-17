-- Clear all existing fighter data (names, records, last 5 fights, matches)
-- Fighter images on disk (static/images/fighters) and the fighter_image_url
-- column definition are untouched by this migration.

DELETE FROM matches;
DELETE FROM fighters;
