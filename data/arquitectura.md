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

### Prerequisitos

- Docker
- Docker Compose

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
