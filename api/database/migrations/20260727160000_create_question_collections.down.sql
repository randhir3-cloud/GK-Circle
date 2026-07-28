-- +migrate Down
-- EXAM-P4-T01: drop Question Collection entities.

DROP TABLE IF EXISTS question_collection_members;
DROP TABLE IF EXISTS question_collections;
