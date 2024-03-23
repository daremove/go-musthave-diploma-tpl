CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');

CREATE TABLE orders (
    id          text PRIMARY KEY,
    user_id     uuid REFERENCES users NOT NULL,
    uploaded_at timestamp NOT NULL DEFAULT current_timestamp
);

CREATE TABLE accruals (
    id      text PRIMARY KEY REFERENCES orders,
    status  order_status NOT NULL DEFAULT 'NEW',
    accrual numeric(15, 2) NOT NULL DEFAULT 0
);
