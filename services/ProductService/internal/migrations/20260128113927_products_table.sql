-- +goose Up
-- +goose StatementBegin
create table products (
    id serial primary key,
    name varchar(100) not null,
    short_description varchar(150),
    description text not null,
    price float not null,
    discount_type varchar(50),
    discount_value float,
    discount_start_date timestamp,
    discount_end_date timestamp,
    image_url varchar(255),
    quantity int not null,
    created_at timestamp with time zone default current_timestamp,
    updated_at timestamp with time zone default current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table products;
-- +goose StatementEnd
