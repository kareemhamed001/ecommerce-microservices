-- +goose Up
-- +goose StatementBegin
create table order_items (
    id serial primary key,
    order_id int not null references orders(id) on delete cascade,
    product_id int not null,
    quantity int not null,
    unit_price float not null,
    total_price float not null,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table order_items;
-- +goose StatementEnd