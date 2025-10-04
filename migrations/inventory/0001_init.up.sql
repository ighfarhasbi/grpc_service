CREATE TYPE reservation_status AS ENUM ('reserved', 'confirmed', 'canceled');

CREATE TABLE public.products (
    products_id UUID PRIMARY KEY,
    sku VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    stock INT NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE public.reservations (
    reservations_id UUID PRIMARY KEY,
    products_id UUID NOT NULL,
    orders_id UUID NOT NULL,
    qty INT NOT NULL,
    status reservation_status NOT NULL DEFAULT 'reserved', 
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    CONSTRAINT reservations_products_id_fkey FOREIGN KEY (products_id) REFERENCES public.products(products_id)
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
CREATE TRIGGER set_timestamp_products
BEFORE UPDATE ON public.products
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Trigger for reservations table
CREATE TRIGGER set_timestamp_reservations
BEFORE UPDATE ON public.reservations
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
