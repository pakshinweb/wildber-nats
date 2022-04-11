-- +migrate Up
CREATE TABLE IF NOT EXISTS public.order
(
    id SERIAL PRIMARY KEY,
    data jsonb default '{}' NOT NULL
);

create index order_data_gin_idx on public.order using gin(data);

-- +migrate Down
DROP TABLE public.order;