BEGIN;

create table groups
(
	id bigint not null
		constraint groups_pkey
			primary key,
	greeting varchar(1000),
	questions jsonb,
	ban_url boolean,
	ban_question boolean
);

comment on table groups is 'Group data.';

comment on column groups.id is 'id.';

comment on column groups.greeting is 'Greeting.';

comment on column groups.questions is 'Questions and answers.';

comment on column groups.ban_url is 'Ban by postings urls and forwards.';

comment on column groups.ban_question is 'Ban by question.';

alter table groups owner to irma;

COMMIT;
