UPDATE releases SET image_repository = '' WHERE image_repository IS NULL;
UPDATE releases SET image_tag = '' WHERE image_tag IS NULL;
UPDATE releases SET image_digest = '' WHERE image_digest IS NULL;

ALTER TABLE releases ALTER COLUMN image_repository SET NOT NULL;
ALTER TABLE releases ALTER COLUMN image_tag SET NOT NULL;
ALTER TABLE releases ALTER COLUMN image_digest SET NOT NULL;
