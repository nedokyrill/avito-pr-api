create table if not exists users (
    id uuid primary key default gen_random_uuid(),
    name varchar(255) not null,
    team_id uuid references teams(id) on delete cascade,
    is_active boolean not null default true,
    created_at timestamp default now()
);

create index idx_users_team_active on users(team_id, is_active);
create index idx_users_team_name on users(team_id);