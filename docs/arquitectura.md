# 🏗️ Arquitectura del Sistema

## Componentes

El sistema de gestión de almacenamiento de noticias se basa en una arquitectura de microservicios asíncrona:

* **PostgreSQL**: Base de datos relacional utilizada para el **almacenamiento persistente** y como fuente única de verdad (Single Source of Truth).
* **Elasticsearch**: Motor de búsqueda de texto completo utilizado para **consultas rápidas** y búsquedas complejas en el contenido de las noticias.
* **RabbitMQ**: Broker de mensajes que desacopla los servicios y gestiona el **procesamiento asíncrono** de las noticias.
* **API Ingestion**: Microservicio principal que expone una API REST para recibir noticias del cliente.
* **Worker Indexer**: Microservicio que procesa mensajes de `RabbitMQ` y los guarda en la base de datos **PostgreSQL**.
* **Worker Sync**: Microservicio que procesa eventos de actualización de PostgreSQL y **sincroniza** la información a **Elasticsearch**.
* **Migrator / ES-Init**: Servicios de inicialización que garantizan que el esquema de PostgreSQL y el índice de Elasticsearch estén listos antes de que arranquen los workers.

## Flujo de Datos

El flujo de una noticia desde la ingesta hasta su almacenamiento y búsqueda es el siguiente:

1. **Ingesta**: El `Cliente` envía la noticia por HTTP a la **API Ingestion**.
2. **Encolado**: La **API Ingestion** pone el mensaje en la cola `ingestion_queue` de RabbitMQ.
3. **Indexación**: El **Worker Indexer** lee de `ingestion_queue` y escribe la noticia en **PostgreSQL**.
4. **Sincronización (Asíncrona)**: Después de la escritura, se genera un evento que se envía a la cola `sync_queue` de RabbitMQ.
5. **Búsqueda**: El **Worker Sync** lee de `sync_queue` y escribe o actualiza el documento en **Elasticsearch**.

```
Cliente → API Ingestion → RabbitMQ (ingestion_queue)
                              ↓
                        Worker Indexer → PostgreSQL
                              ↓
                    RabbitMQ (sync_queue)
                              ↓
                        Worker Sync → Elasticsearch
```

### Componentes

- **PostgreSQL**: Base de datos relacional para almacenamiento persistente
- **Elasticsearch**: Motor de búsqueda para consultas rápidas
- **RabbitMQ**: Cola de mensajes para procesamiento asíncrono
- **API Ingestion**: API REST para recibir noticias
- **Worker Indexer**: Procesa noticias y las guarda en PostgreSQL
- **Worker Sync**: Sincroniza noticias de PostgreSQL a Elasticsearch

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
