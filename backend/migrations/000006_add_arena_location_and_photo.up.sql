ALTER TABLE arenas
    ADD COLUMN latitude DOUBLE PRECISION,
    ADD COLUMN longitude DOUBLE PRECISION,
    ADD COLUMN photo_url TEXT,
    ADD CONSTRAINT arenas_coordinates_pair_check
        CHECK ((latitude IS NULL) = (longitude IS NULL)),
    ADD CONSTRAINT arenas_latitude_range_check
        CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90),
    ADD CONSTRAINT arenas_longitude_range_check
        CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180);
