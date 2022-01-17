BEGIN;

ALTER TABLE public.groups ALTER COLUMN ban_timeout TYPE smallint USING ban_timeout::smallint;

COMMIT;
