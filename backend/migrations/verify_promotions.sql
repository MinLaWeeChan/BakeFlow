-- Quick verification script for promotions table
-- Run this to check if everything is set up correctly

-- 1. Check table exists and has correct structure
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'promotions'
ORDER BY ordinal_position;

-- 2. Check constraints
SELECT 
    constraint_name,
    constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'promotions';

-- 3. Check indexes
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'promotions';

-- 4. Check trigger
SELECT 
    trigger_name,
    event_manipulation,
    action_statement
FROM information_schema.triggers
WHERE event_object_table = 'promotions';

-- 5. Check if example data exists
SELECT id, name, type, active, priority FROM promotions;
