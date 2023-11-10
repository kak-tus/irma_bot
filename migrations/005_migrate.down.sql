BEGIN;

ALTER TABLE public.groups DROP COLUMN ban_emojii_count;

COMMIT;
