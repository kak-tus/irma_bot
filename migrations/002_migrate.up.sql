BEGIN;

ALTER TABLE public.groups ADD COLUMN ban_timeout smallint;

COMMENT ON COLUMN public.groups.ban_timeout IS 'Ban delay after question';

COMMIT;
