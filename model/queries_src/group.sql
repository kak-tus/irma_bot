-- name: GetGroup :one
SELECT
  ban_question,
  ban_url,
  greeting,
  questions,
  ban_timeout,
  ignore_domain
FROM public.groups
WHERE id = $1;

-- name: CreateOrUpdateGroup :exec
INSERT INTO groups
  (id, greeting, questions, ban_url, ban_question, ban_timeout, ignore_domain)
  VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
  (greeting, questions, ban_url, ban_question, ban_timeout, ignore_domain) =
  ROW(
    EXCLUDED.greeting, EXCLUDED.questions, EXCLUDED.ban_url,
    EXCLUDED.ban_question, EXCLUDED.ban_timeout, EXCLUDED.ignore_domain
    );
