create type pr_status as enum ('OPEN', 'MERGED');

create table if not exists pull_requests (
    id uuid primary key default gen_random_uuid(),
    name text not null,
    author_id uuid references users(id) on delete cascade,
    status pr_status not null default 'OPEN',
    need_more_reviewers boolean not null default false,
    created_at timestamp default now(),
    merged_at timestamp
);

create index idx_pr_author on pull_requests(author_id);
create index idx_pr_status on pull_requests(status);