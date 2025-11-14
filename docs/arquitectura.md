# 🏗️ Arquitectura del Sistema

El **Data Storage Manager** opera bajo una arquitectura de microservicios, diseñada para el procesamiento y almacenamiento asíncrono de grandes volúmenes de datos de noticias.

## 1. 🌐 Componentes Principales

El sistema se compone de servicios de aplicación desacoplados y una infraestructura de datos robusta:

### Aplicaciones (Microservicios)

* **API Ingestion**: API REST que actúa como la puerta de entrada, recibiendo el formato nativo del scraper.
* **Worker Indexer**: Procesa las noticias recibidas, incluyendo la tarea crítica de parsear fechas españolas y guardar la información en **PostgreSQL**.
* **Worker Sync**: Responsable de sincronizar las noticias desde **PostgreSQL** a **Elasticsearch** para habilitar la búsqueda rápida.

### Infraestructura de Datos

* **PostgreSQL**: Base de datos relacional utilizada como la **fuente única de verdad** para el almacenamiento persistente.
* **Elasticsearch**: Motor de búsqueda optimizado para consultas rápidas y de texto completo.
* **RabbitMQ**: Cola de mensajes que garantiza el procesamiento asíncrono y la resiliencia entre los servicios.

---

## 2. 🌊 Flujo de Datos (Asíncrono)

El procesamiento de una noticia es completamente asíncrono, lo que garantiza que la API de ingesta responda rápidamente y el procesamiento se realice sin bloquear al cliente.
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
---

## 4. 🗃️ Detalle de Componentes Clave

### API Ingestion - Modelo de Datos

La API recibe el formato nativo del scraper y lo mapea al siguiente DTO (Data Transfer Object) en Go:

```go
type NewsRequest struct {
    URL            string   `json:"url"`
    Titulo         string   `json:"titulo"`
    Fecha          string   `json:"fecha"`           // Formato español
    Tags           []string `json:"tags"`
    Autor          string   `json:"autor"`
    DescAutor      string   `json:"desc_autor"`
    Abstract       string   `json:"abstract"`
    Cuerpo         string   `json:"cuerpo"`
    Multimedia     []string `json:"multimedia"`
    TipoMultimedia string   `json:"tipo_multimedia"`
}
```
### Worker Indexer (Logica de Procesamiento)

Este worker contiene la logica clave de transformacion y validacion:
- **Parser de fechas**: Tiene un parser de fechas españolas integrado.
- **Extracción de Metadata**: Realiza la auto-extracción del medio desde la URL.
- **Gestión de Tags**: Se encarga de la creación automática de tags y su relación many-to-many.
- **Detección de Duplicados**: Utiliza un Hash SHA256 de la URL normalizada para prevenir ingestas duplicadas.