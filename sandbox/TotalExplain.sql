-- 1. Setup parameters (adjust values to test different scenarios)
-- Test Case: Specific User + Service + Date Range
EXPLAIN ANALYZE
SELECT COALESCE(SUM(price), 0)
FROM subscriptions
WHERE ('60601fee-2bf1-4721-ae6f-7636e79a0cba'::uuid IS NULL OR user_id = '60601fee-2bf1-4721-ae6f-7636e79a0cba')
  AND ('Yandex Plus' = '' OR service_name = 'Yandex Plus')
  AND start_date <= '2025-12-31'
  AND (end_date IS NULL OR end_date >= '2025-01-01');
