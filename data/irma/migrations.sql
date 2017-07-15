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
