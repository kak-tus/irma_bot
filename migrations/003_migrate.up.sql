BEGIN;

ALTER TABLE public.groups ALTER COLUMN ban_timeout TYPE integer;

COMMIT;
