create table if not exists pr_reviewers (
    pull_request_id varchar(255) not null references pull_requests(id) on delete cascade,
    reviewer_id varchar(255) not null references users(id) on delete cascade,
    assigned_at timestamp default now(),
    primary key (pull_request_id, reviewer_id)
);

-- сделал только один индекс на ревьюера, так как с пр айди справится праймари кей
create index idx_pr_reviewer on pr_reviewers(reviewer_id);