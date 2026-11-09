ALTER TABLE arenas
    DROP CONSTRAINT arenas_longitude_range_check,
    DROP CONSTRAINT arenas_latitude_range_check,
    DROP CONSTRAINT arenas_coordinates_pair_check,
    DROP COLUMN photo_url,
    DROP COLUMN longitude,
    DROP COLUMN latitude;
