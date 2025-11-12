create table if not exists teams (
    id uuid primary key default gen_random_uuid(),
    name varchar(100) not null,
    created_at timestamp default now()
);

create unique index idx_teams_name on teams(name);