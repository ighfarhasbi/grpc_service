CREATE TYPE order_status AS ENUM ('pending', 'confirmed', 'canceled');

CREATE TABLE public.orders (
    orders_id UUID PRIMARY KEY,
    users_id UUID NOT NULL,
    total_amount NUMERIC(11,2) NOT NULL,
    status order_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE public.order_items (
    order_items_id UUID PRIMARY KEY,
    orders_id UUID NOT NULL,
    products_id UUID NOT NULL,
    qty INT NOT NULL,
    unit_price NUMERIC(10,2) NOT NULL,
    total_price NUMERIC(11,2) GENERATED ALWAYS AS (qty * unit_price) STORED,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT order_items_orders_id_fkey FOREIGN KEY (orders_id) REFERENCES public.orders(orders_id)
);

-- Function to auto-update updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for products table
CREATE TRIGGER set_timestamp_orders
BEFORE UPDATE ON public.orders
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Trigger for reservations table
CREATE TRIGGER set_timestamp_order_items
BEFORE UPDATE ON public.order_items
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
