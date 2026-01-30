-- +goose Up
-- +goose StatementBegin
create table orders (
    id serial primary key,
    user_id int not null,
    shipping_cost float not null default 0,
    shipping_duration_days int not null default 0,
    discount float not null default 0,
    total float not null default 0,
    status varchar(20) not null default 'pending',
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table orders;
-- +goose StatementEnd