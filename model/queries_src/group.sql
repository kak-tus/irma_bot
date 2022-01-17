-- name: GetGroup :one
SELECT
  ban_question,
  ban_url,
  greeting,
  questions,
  ban_timeout
FROM public.groups
WHERE id = $1;

-- name: CreateOrUpdateGroupQuestions :exec
INSERT INTO groups
  (id, greeting, questions)
  VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE SET
  (greeting, questions) =
  ROW(EXCLUDED.greeting, EXCLUDED.questions);

-- name: CreateOrUpdateGroupParameters :exec
INSERT INTO groups
  (id, ban_url, ban_question, ban_timeout)
  VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
  (ban_url, ban_question, ban_timeout) =
  ROW(EXCLUDED.ban_url, EXCLUDED.ban_question, EXCLUDED.ban_timeout);
