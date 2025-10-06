DROP TRIGGER IF EXISTS set_timestamp_orders ON public.orders;
DROP TRIGGER IF EXISTS set_timestamp_order_items ON public.order_items; 
DROP TABLE IF EXISTS public.orders;
DROP TABLE IF EXISTS public.order_items;
DROP TYPE IF EXISTS order_status;
DROP FUNCTION IF EXISTS update_updated_at_column();
