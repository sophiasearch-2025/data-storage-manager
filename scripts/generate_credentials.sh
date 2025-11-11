#!/bin/bash

# ==============================================
# Script para generar credenciales seguras
# ==============================================

set -e

# Colores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}==============================================
Generador de Credenciales Seguras
==============================================${NC}"

# Función para generar password aleatorio
generate_password() {
    # Genera password de 32 caracteres con letras, números y símbolos seguros
    openssl rand -base64 32 | tr -d "=+/" | cut -c1-32
}

# Verificar si ya existe .env
if [ -f ".env" ]; then
    echo -e "${YELLOW}⚠️  El archivo .env ya existe.${NC}"
    read -p "¿Deseas sobrescribirlo? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}❌ Operación cancelada${NC}"
        exit 1
    fi

    # Backup del .env existente
    BACKUP_FILE=".env.backup.$(date +%Y%m%d_%H%M%S)"
    cp .env "$BACKUP_FILE"
    echo -e "${GREEN}✓ Backup creado: $BACKUP_FILE${NC}"
fi

# Generar passwords
echo -e "\n${GREEN}🔐 Generando credenciales seguras...${NC}"

POSTGRES_PASSWORD=$(generate_password)
POSTGRES_WORKER_PASSWORD=$(generate_password)
POSTGRES_READONLY_PASSWORD=$(generate_password)

echo -e "${GREEN}✓ Credenciales generadas${NC}"

# Crear archivo .env desde .env.example
if [ ! -f ".env.example" ]; then
    echo -e "${RED}❌ Error: .env.example no encontrado${NC}"
    exit 1
fi

# Copiar .env.example y reemplazar valores
cp .env.example .env

# Reemplazar passwords en .env (compatible con macOS y Linux)
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    sed -i '' "s/POSTGRES_WORKER_PASSWORD=.*/POSTGRES_WORKER_PASSWORD=${POSTGRES_WORKER_PASSWORD}/" .env
    sed -i '' "s/POSTGRES_READONLY_PASSWORD=.*/POSTGRES_READONLY_PASSWORD=${POSTGRES_READONLY_PASSWORD}/" .env
else
    # Linux
    sed -i "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${POSTGRES_PASSWORD}/" .env
    sed -i "s/POSTGRES_WORKER_PASSWORD=.*/POSTGRES_WORKER_PASSWORD=${POSTGRES_WORKER_PASSWORD}/" .env
    sed -i "s/POSTGRES_READONLY_PASSWORD=.*/POSTGRES_READONLY_PASSWORD=${POSTGRES_READONLY_PASSWORD}/" .env
fi

echo -e "\n${GREEN}==============================================
✅ Archivo .env creado exitosamente
==============================================${NC}"

echo -e "\n${YELLOW}📋 Credenciales generadas:${NC}"
echo -e "PostgreSQL Admin:      ${GREEN}${POSTGRES_PASSWORD}${NC}"
echo -e "PostgreSQL Worker:     ${GREEN}${POSTGRES_WORKER_PASSWORD}${NC}"
echo -e "PostgreSQL Read-Only:  ${GREEN}${POSTGRES_READONLY_PASSWORD}${NC}"

echo -e "\n${YELLOW}⚠️  IMPORTANTE:${NC}"
echo "1. Guarda estas credenciales en un lugar seguro"
echo "2. El archivo .env NO debe ser commiteado a git"
echo "3. Verifica que .env esté en .gitignore"

# Verificar .gitignore
if grep -q "^\.env$" .gitignore 2>/dev/null; then
    echo -e "${GREEN}✓ .env está en .gitignore${NC}"
else
    echo -e "${RED}❌ ADVERTENCIA: .env NO está en .gitignore${NC}"
    echo -e "${YELLOW}   Ejecuta: echo '.env' >> .gitignore${NC}"
fi

echo -e "\n${GREEN}==============================================
Próximos pasos:
==============================================${NC}"
echo "1. Revisar y ajustar configuraciones en .env"
echo "2. Ejecutar: docker-compose down -v"
echo "3. Ejecutar: docker-compose up -d"
echo "4. Las nuevas credenciales se aplicarán"

echo -e "\n${GREEN}✓ Listo!${NC}\n"
