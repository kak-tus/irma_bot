BEGIN;

ALTER TABLE public.groups ADD COLUMN ban_emojii_count integer;

COMMIT;
