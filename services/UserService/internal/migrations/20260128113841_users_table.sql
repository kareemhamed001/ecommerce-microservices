-- +goose Up
-- +goose StatementBegin
create table users (
    id serial primary key,
    name varchar(50) not null unique,
    email varchar(100) not null unique,
    password varchar(255) not null,
    role varchar(20) not null default 'customer',
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table users;
-- +goose StatementEnd
