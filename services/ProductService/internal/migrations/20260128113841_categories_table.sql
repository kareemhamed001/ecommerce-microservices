-- +goose Up
-- +goose StatementBegin
create table categories (
    id serial primary key,
    name varchar(100) not null,
    description text,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table categories;
-- +goose StatementEnd
