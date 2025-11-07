# NewsPress - Data Storage Manager

Sistema de gestión de almacenamiento de noticias con arquitectura de microservicios.

## Arquitectura

### Componentes

- **PostgreSQL**: Base de datos relacional para almacenamiento persistente
- **Elasticsearch**: Motor de búsqueda para consultas rápidas
- **RabbitMQ**: Cola de mensajes para procesamiento asíncrono
- **API Ingestion**: API REST para recibir noticias
- **Worker Indexer**: Procesa noticias y las guarda en PostgreSQL
- **Worker Sync**: Sincroniza noticias de PostgreSQL a Elasticsearch

### Flujo de Datos

```
Cliente → API Ingestion → RabbitMQ (ingestion_queue)
                              ↓
                        Worker Indexer → PostgreSQL
                              ↓
                    RabbitMQ (sync_queue)
                              ↓
                        Worker Sync → Elasticsearch
```

## Inicio Rápido

### Prerequisitos

- Docker
- Docker Compose

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

## Desarrollo

### Estructura del Proyecto

```
.
├── api-ingestion/          # API REST para ingesta
├── workers/
│   ├── common/             # Configuración compartida
│   ├── indexer/            # Worker de indexación a Postgres
│   └── sync/               # Worker de sincronización a ES
├── database/
│   └── migrations/         # Migraciones SQL
├── elasticsearch/
│   └── init/               # Scripts de inicialización
└── docker-compose.yml      # Orquestación de servicios
```

### Logs

```bash
# Ver logs de un servicio específico
docker-compose logs -f api-ingestion
docker-compose logs -f worker-indexer
docker-compose logs -f worker-sync

# Ver logs de todos los servicios
docker-compose logs -f
```

### Detener el sistema

```bash
# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (CUIDADO: elimina datos)
docker-compose down -v
```

## Pendientes

- [ ] API de consultas (query API)
- [ ] Tests unitarios y de integración
- [ ] Logging estructurado
- [ ] Métricas y monitoreo
- [ ] Autenticación y autorización
- [ ] Dead letter queue para mensajes fallidos