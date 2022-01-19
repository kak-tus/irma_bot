-- name: GetGroup :one
SELECT
  ban_question,
  ban_url,
  greeting,
  questions,
  ban_timeout
FROM public.groups
WHERE id = $1;

-- name: CreateOrUpdateGroup :exec
INSERT INTO groups
  (id, greeting, questions, ban_url, ban_question, ban_timeout)
  VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
  (greeting, questions, ban_url, ban_question, ban_timeout) =
  ROW(EXCLUDED.greeting, EXCLUDED.questions, EXCLUDED.ban_url, EXCLUDED.ban_question, EXCLUDED.ban_timeout);
