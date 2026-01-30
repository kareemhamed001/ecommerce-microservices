-- +goose Up
-- +goose StatementBegin
create table addresses(
    id serial primary key,
    user_id integer not null references users(id) on delete cascade,
    country varchar(50) not null,
    city varchar(50) not null,
    state varchar(50) not null,
    street varchar(100) not null,
    zip_code varchar(20) null,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table addresses;
-- +goose StatementEnd
