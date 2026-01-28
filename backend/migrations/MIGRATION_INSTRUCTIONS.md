# Migration Instructions - Promotions System

## Situation
You ran the first migration (`008_create_promotions.sql`) and now want to use the updated version (`011_create_promotions.sql`) which has:
- Better constraints
- Trigger for updated_at
- More indexes
- Example data

## Step-by-Step Migration

### Option 1: Rollback and Recreate (Recommended if no important data)

```bash
cd backend

# 1. Rollback the old migration
psql "$DATABASE_URL" -f migrations/008_rollback_promotions.sql

# 2. Run the new migration
psql "$DATABASE_URL" -f migrations/011_create_promotions.sql
```

### Option 2: Keep Data and Upgrade (If you have promotions you want to keep)

```bash
cd backend

# 1. Backup your data (optional but recommended)
psql "$DATABASE_URL" -c "\COPY promotions TO '/tmp/promotions_backup.csv' CSV HEADER;"

# 2. Add missing columns and constraints
psql "$DATABASE_URL" <<EOF
-- Add constraint if it doesn't exist
DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_dates'
    ) THEN
        ALTER TABLE promotions ADD CONSTRAINT check_dates CHECK (end_at > start_at);
    END IF;
END \$\$;

-- Add trigger if it doesn't exist
DROP TRIGGER IF EXISTS update_promotions_updated_at ON promotions;
CREATE TRIGGER update_promotions_updated_at
    BEFORE UPDATE ON promotions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add missing indexes
CREATE INDEX IF NOT EXISTS idx_promotions_priority ON promotions(priority DESC);
CREATE INDEX IF NOT EXISTS idx_promotions_type ON promotions(type);
EOF

# 3. Verify the upgrade
psql "$DATABASE_URL" -c "\d promotions"
```

### Option 3: Manual SQL (If you prefer step-by-step)

```sql
-- Connect to your database
\c your_database_name

-- 1. Check current table structure
\d promotions

-- 2. Drop old table (WARNING: This deletes all data!)
DROP TABLE IF EXISTS promotions CASCADE;

-- 3. Run the new migration
\i migrations/011_create_promotions.sql

-- 4. Verify
\d promotions
SELECT * FROM promotions;
```

## Verify Migration Success

After running the migration, verify it worked:

```bash
psql "$DATABASE_URL" -c "
SELECT 
    table_name,
    column_name,
    data_type,
    is_nullable
FROM information_schema.columns
WHERE table_name = 'promotions'
ORDER BY ordinal_position;
"
```

You should see:
- All columns present
- `check_dates` constraint exists
- Trigger `update_promotions_updated_at` exists
- All indexes created

## Check Indexes

```bash
psql "$DATABASE_URL" -c "\d promotions"
```

Should show all indexes:
- `idx_promotions_active`
- `idx_promotions_dates`
- `idx_promotions_priority`
- `idx_promotions_type`

## Check Trigger

```bash
psql "$DATABASE_URL" -c "
SELECT trigger_name, event_manipulation, event_object_table
FROM information_schema.triggers
WHERE event_object_table = 'promotions';
"
```

Should show: `update_promotions_updated_at`

## Differences Between Versions

| Feature | Old (008) | New (011) |
|---------|-----------|-----------|
| Name column | TEXT | TEXT (same) |
| Default active | TRUE | TRUE (same) |
| Date constraint | ❌ | ✅ check_dates |
| Trigger | ❌ | ✅ updated_at |
| Indexes | 3 basic | 4 (includes type) |
| Example data | ❌ | ✅ 2 examples |

## Troubleshooting

### Error: "function update_updated_at_column() does not exist"
This function should exist from previous migrations. If not, create it:

```sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS \$\$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
\$\$ LANGUAGE plpgsql;
```

### Error: "relation promotions already exists"
You need to drop the old table first:
```bash
psql "$DATABASE_URL" -f migrations/008_rollback_promotions.sql
```

### Error: "constraint check_dates already exists"
The constraint already exists, which is fine. The migration uses `IF NOT EXISTS` for safety.

## After Migration

1. Test the API:
   ```bash
   curl http://localhost:8080/promotions/active
   ```

2. Create a test promotion:
   ```sql
   INSERT INTO promotions (name, type, rules, active, start_at, end_at, priority)
   VALUES (
     'Test Promotion',
     'PERCENT_OFF',
     '{"percent": 10, "productIds": []}'::jsonb,
     true,
     NOW(),
     NOW() + INTERVAL '7 days',
     5
   );
   ```

3. Verify it appears in API:
   ```bash
   curl http://localhost:8080/promotions/active
   ```
