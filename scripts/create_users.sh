#!/bin/bash
set -e

# Colores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}==============================================
Creación de Usuarios PostgreSQL
==============================================${NC}"

# Verificar que existe .env
if [ ! -f ".env" ]; then
    echo -e "${RED}❌ Error: Archivo .env no encontrado${NC}"
    echo -e "${YELLOW}   Ejecuta primero: bash scripts/generate_credentials.sh${NC}"
    exit 1
fi

# Cargar variables del .env
source .env

# Verificar que las variables necesarias existan
if [ -z "$POSTGRES_WORKER_PASSWORD" ] || [ -z "$POSTGRES_READONLY_PASSWORD" ]; then
    echo -e "${RED}❌ Error: Variables POSTGRES_WORKER_PASSWORD o POSTGRES_READONLY_PASSWORD no definidas en .env${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Variables de entorno cargadas${NC}"

# Verificar que PostgreSQL esté corriendo
if ! docker ps | grep -q "data-storage-manager-postgres"; then
    echo -e "${RED}❌ Error: PostgreSQL no está corriendo${NC}"
    echo -e "${YELLOW}   Ejecuta: docker-compose up -d postgres${NC}"
    exit 1
fi

echo -e "${GREEN}✓ PostgreSQL está corriendo${NC}"

# Esperar a que PostgreSQL esté listo
echo -e "${YELLOW}⏳ Esperando a que PostgreSQL esté listo...${NC}"
sleep 3

# Copiar el script SQL y reemplazar placeholders
TEMP_SQL="/tmp/create_users_temp.sql"
cp database/migrations/postgres/002_create_users.sql "$TEMP_SQL"

# Reemplazar placeholders con passwords reales
sed -i.bak "s/WORKER_PASSWORD_PLACEHOLDER/${POSTGRES_WORKER_PASSWORD}/g" "$TEMP_SQL"
sed -i.bak "s/READONLY_PASSWORD_PLACEHOLDER/${POSTGRES_READONLY_PASSWORD}/g" "$TEMP_SQL"

echo -e "${GREEN} Script SQL preparado${NC}"

# Ejecutar el script en PostgreSQL
echo -e "${YELLOW} Creando usuarios en PostgreSQL...${NC}"

docker exec -i data-storage-manager-postgres psql -U postgres -d newspress < "$TEMP_SQL"

# Limpiar archivo temporal
rm -f "$TEMP_SQL" "$TEMP_SQL.bak"

echo -e "\n${GREEN}==============================================
Usuarios creados exitosamente
==============================================${NC}"

echo -e "\n${YELLOW}📋 Resumen:${NC}"
echo -e "${GREEN}✓${NC} Usuario: newspress_worker"
echo -e "  - Permisos: SELECT, INSERT, UPDATE"
echo -e "  - Usado por: worker-indexer, worker-sync"
echo ""
echo -e "${GREEN}✓${NC} Usuario: newspress_readonly"
echo -e "  - Permisos: SELECT (solo lectura)"
echo -e "  - Usado por: API de consultas"

echo -e "\n${YELLOW}⚠️  Próximos pasos:${NC}"
echo "1. Actualizar configuración de workers para usar newspress_worker"
echo "2. Reiniciar workers: docker-compose restart worker-indexer worker-sync"
echo "3. Verificar logs: docker logs data-storage-manager-worker-indexer"

echo -e "\n${GREEN}✓ Listo!${NC}\n"
