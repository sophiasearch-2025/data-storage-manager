### Levantar el sistema

```bash
# 1. Clonar variables de entorno
cp .env.example .env

# 2. Iniciar todos los servicios
docker-compose up -d

# 3. Verificar que todos los servicios estén corriendo
docker-compose ps
```

### Orden de Inicio

El sistema se inicia automáticamente en el orden correcto:

1. Infraestructura (postgres, elasticsearch, rabbitmq)
2. Migrator (ejecuta schema SQL)
3. ES-Init (crea índice en Elasticsearch)
4. API y Workers

### Probar el sistema

```bash
# Enviar una noticia de prueba
curl -X POST http://localhost:8080/api/v1/news \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/noticia1",
    "title": "Noticia de prueba",
    "content": "Este es el contenido de la noticia de prueba",
    "abstract": "Resumen de la noticia",
    "author": "Autor de Prueba",
    "media_outlet": "El Medio",
    "country": "chile",
    "published_date": "2025-11-07T10:00:00Z"
  }'

# Verificar en PostgreSQL
docker exec data-storage-manager-postgres psql -U postgres -d newspress -c "SELECT id, title FROM news;"

# Verificar en Elasticsearch
curl -X GET "http://localhost:9200/news/_search?pretty"
```

## Endpoints Disponibles

### API Ingestion

- `GET /health` - Health check
- `POST /api/v1/news` - Ingestar nueva noticia

## Acceso a Servicios

- **API Ingestion**: http://localhost:8080
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **Elasticsearch**: http://localhost:9200
- **PostgreSQL**: localhost:5432 (postgres/postgres123)
