BEGIN;

ALTER TABLE public.groups DROP COLUMN ban_timeout;

COMMIT;
