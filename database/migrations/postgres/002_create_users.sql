
-- Terminar conexiones activas de los usuarios si existen
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE usename IN ('newspress_worker', 'newspress_readonly')
  AND pid <> pg_backend_pid();

-- Eliminar usuarios si ya existen (para re-ejecutar script)
DROP USER IF EXISTS newspress_worker;
DROP USER IF EXISTS newspress_readonly;

-- ==============================================
-- 1. Usuario Worker (Read/Write)
-- ==============================================
-- Usado por: worker-indexer, worker-sync
-- Permisos: SELECT, INSERT, UPDATE en todas las tablas

-- Crear usuario worker
CREATE USER newspress_worker WITH PASSWORD 'WORKER_PASSWORD_PLACEHOLDER';

-- Otorgar conexión a la base de datos
GRANT CONNECT ON DATABASE newspress TO newspress_worker;

-- Otorgar uso del schema public
GRANT USAGE ON SCHEMA public TO newspress_worker;

-- Otorgar permisos de lectura y escritura en tablas existentes
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA public TO newspress_worker;

-- Otorgar permisos en secuencias (para IDs autoincrementales)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO newspress_worker;

-- Otorgar permisos automáticos en tablas futuras
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE ON TABLES TO newspress_worker;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO newspress_worker;

-- Log
DO $$
BEGIN
    RAISE NOTICE '✓ Usuario newspress_worker creado con permisos R/W';
END $$;

-- ==============================================
-- 2. Usuario Read-Only (Solo Lectura)
-- ==============================================
-- Usado por: API de consultas (futuro)
-- Permisos: Solo SELECT en todas las tablas

-- Crear usuario readonly
CREATE USER newspress_readonly WITH PASSWORD 'READONLY_PASSWORD_PLACEHOLDER';

-- Otorgar conexión a la base de datos
GRANT CONNECT ON DATABASE newspress TO newspress_readonly;

-- Otorgar uso del schema public
GRANT USAGE ON SCHEMA public TO newspress_readonly;

-- Otorgar solo SELECT en tablas existentes
GRANT SELECT ON ALL TABLES IN SCHEMA public TO newspress_readonly;

-- Otorgar permisos automáticos en tablas futuras
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO newspress_readonly;

-- Log
DO $$
BEGIN
    RAISE NOTICE '✓ Usuario newspress_readonly creado con permisos R/O';
END $$;

-- ==============================================
-- Verificación de Permisos
-- ==============================================

-- Mostrar permisos de newspress_worker
DO $$
DECLARE
    r RECORD;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '================================================';
    RAISE NOTICE 'Permisos de newspress_worker:';
    RAISE NOTICE '================================================';

    FOR r IN
        SELECT tablename,
               has_table_privilege('newspress_worker', schemaname||'.'||tablename, 'SELECT') as can_select,
               has_table_privilege('newspress_worker', schemaname||'.'||tablename, 'INSERT') as can_insert,
               has_table_privilege('newspress_worker', schemaname||'.'||tablename, 'UPDATE') as can_update,
               has_table_privilege('newspress_worker', schemaname||'.'||tablename, 'DELETE') as can_delete
        FROM pg_tables
        WHERE schemaname = 'public'
        ORDER BY tablename
    LOOP
        RAISE NOTICE 'Table: % | SELECT: % | INSERT: % | UPDATE: % | DELETE: %',
            r.tablename, r.can_select, r.can_insert, r.can_update, r.can_delete;
    END LOOP;
END $$;

-- Mostrar permisos de newspress_readonly
DO $$
DECLARE
    r RECORD;
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '================================================';
    RAISE NOTICE 'Permisos de newspress_readonly:';
    RAISE NOTICE '================================================';

    FOR r IN
        SELECT tablename,
               has_table_privilege('newspress_readonly', schemaname||'.'||tablename, 'SELECT') as can_select,
               has_table_privilege('newspress_readonly', schemaname||'.'||tablename, 'INSERT') as can_insert,
               has_table_privilege('newspress_readonly', schemaname||'.'||tablename, 'UPDATE') as can_update,
               has_table_privilege('newspress_readonly', schemaname||'.'||tablename, 'DELETE') as can_delete
        FROM pg_tables
        WHERE schemaname = 'public'
        ORDER BY tablename
    LOOP
        RAISE NOTICE 'Table: % | SELECT: % | INSERT: % | UPDATE: % | DELETE: %',
            r.tablename, r.can_select, r.can_insert, r.can_update, r.can_delete;
    END LOOP;
END $$;

-- ==============================================
-- Resumen
-- ==============================================

DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '================================================';
    RAISE NOTICE '✅ USUARIOS CREADOS EXITOSAMENTE';
    RAISE NOTICE '================================================';
    RAISE NOTICE '';
    RAISE NOTICE '1. newspress_worker:';
    RAISE NOTICE '   - Usado por: worker-indexer, worker-sync';
    RAISE NOTICE '   - Permisos: SELECT, INSERT, UPDATE';
    RAISE NOTICE '   - NO puede: DELETE';
    RAISE NOTICE '';
    RAISE NOTICE '2. newspress_readonly:';
    RAISE NOTICE '   - Usado por: API de consultas';
    RAISE NOTICE '   - Permisos: SELECT solamente';
    RAISE NOTICE '   - NO puede: INSERT, UPDATE, DELETE';
    RAISE NOTICE '';
    RAISE NOTICE '⚠️  IMPORTANTE:';
    RAISE NOTICE '   - Las contraseñas por defecto deben cambiarse';
    RAISE NOTICE '   - Usar: ALTER USER newspress_worker PASSWORD ''nueva_password'';';
    RAISE NOTICE '';
    RAISE NOTICE '================================================';
END $$;
