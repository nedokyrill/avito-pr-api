create table if not exists users (
    id varchar(255) primary key,
    name varchar(255) not null,
    team_id uuid references teams(id) on delete cascade,
    is_active boolean not null default true,
    created_at timestamp default now()
);

create index idx_users_team_active on users(team_id, is_active);
create index idx_users_team_name on users(team_id);