BEGIN;

ALTER TABLE public.groups DROP COLUMN ignore_domain;

COMMIT;
