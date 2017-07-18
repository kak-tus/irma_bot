-- 1 up

CREATE TABLE groups (
  id bigint NOT NULL,
  greeting varchar(1000) NOT NULL,
  questions jsonb NOT NULL
);

ALTER TABLE groups ADD PRIMARY KEY (id);

COMMENT ON TABLE groups IS 'Group data.';

COMMENT ON COLUMN groups.id IS 'id.';
COMMENT ON COLUMN groups.greeting IS 'Greeting.';
COMMENT ON COLUMN groups.questions IS 'Questions and answers.';

-- 1 down

DROP TABLE groups;

-- 2 up

ALTER TABLE groups ADD COLUMN ban_url boolean DEFAULT TRUE;
ALTER TABLE groups ADD COLUMN ban_question boolean;

COMMENT ON COLUMN groups.ban_url IS 'Ban by postings urls and forwards.';
COMMENT ON COLUMN groups.ban_question IS 'Ban by question.';

ALTER TABLE groups ALTER COLUMN greeting DROP NOT NULL;
ALTER TABLE groups ALTER COLUMN questions DROP NOT NULL;

-- 2 down

ALTER TABLE groups DROP COLUMN ban_url;
ALTER TABLE groups DROP COLUMN ban_question;

ALTER TABLE groups ALTER COLUMN greeting SET NOT NULL;
ALTER TABLE groups ALTER COLUMN questions SET NOT NULL;
