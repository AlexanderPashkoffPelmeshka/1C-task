
create table if not exists cases (
  id          integer primary key autoincrement,
  title       text not null default '',
  description text not null default '',
  status      text not null default 'open'
              check (status in ('open','done','cancelled', 'draft')),
  promised_at text not null,
  due_at      text not null,
  created_at  text not null default (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  updated_at  text not null default (strftime('%Y-%m-%dT%H:%M:%SZ','now')),
  closed_at   text,

  check (status = 'draft' or due_at >= promised_at),
  check (
    (status = 'draft') or
    (status = 'open' and closed_at is null) or
    (status in ('done','cancelled') and closed_at is not null)
  )
);

create index if not exists idx_cases_due    on cases(due_at);
create index if not exists idx_cases_status on cases(status);

create table if not exists comments (
  id         integer primary key autoincrement,
  case_id    integer not null references cases(id) on delete cascade,
  body       text not null,
  created_at text not null default (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);
create index if not exists idx_comments_case on comments(case_id);

create table if not exists drafts (
  scope      text primary key,
  payload    text not null,
  updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%SZ','now'))
);