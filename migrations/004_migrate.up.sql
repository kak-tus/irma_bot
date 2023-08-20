BEGIN;

ALTER TABLE public.groups ADD COLUMN ignore_domain varchar(100) ARRAY[100];

COMMIT;
