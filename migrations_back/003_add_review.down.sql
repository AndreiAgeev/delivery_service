ALTER TABLE couriers
DROP COLUMN IF EXISTS rating,
DROP COLUMN IF EXISTS total_reviews;

DROP INDEX IF EXISTS idx_reviews_courier_id;

DROM TABLE IF EXISTS reviews