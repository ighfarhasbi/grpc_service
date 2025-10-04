DROP TRIGGER IF EXISTS set_timestamp_products ON public.products;
DROP TRIGGER IF EXISTS set_timestamp_reservations ON public.reservations; 
DROP TABLE IF EXISTS public.products;
DROP TABLE IF EXISTS public.reservations;
DROP TYPE IF EXISTS reservation_status;
DROP FUNCTION IF EXISTS update_updated_at_column();
