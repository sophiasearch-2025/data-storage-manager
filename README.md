# NewsPress - Data Storage Manager

Sistema de gestión de almacenamiento de noticias con arquitectura de microservicios.


## Arquitectura

### Componentes

- **PostgreSQL**: Base de datos relacional para almacenamiento persistente
- **Elasticsearch**: Motor de búsqueda para consultas rápidas
- **RabbitMQ**: Cola de mensajes para procesamiento asíncrono
- **API Ingestion**: API REST que recibe el formato nativo del scraper (español)
- **Worker Indexer**: Procesa noticias, parsea fechas españolas y guarda en PostgreSQL
- **Worker Sync**: Sincroniza noticias de PostgreSQL a Elasticsearch

### Flujo de Datos

```
output.json (scraper) → API Ingestion → RabbitMQ (ingestion_queue)
                                             ↓
                                       Worker Indexer → PostgreSQL
                                             ↓
                                   RabbitMQ (sync_queue)
                                             ↓
                                       Worker Sync → Elasticsearch
```

## Formato del Scraper (Nativo)

El sistema acepta directamente el formato del scraper en español:

```json
{
  "url": "https://www.biobiochile.cl/noticias/...",
  "titulo": "Título de la noticia",
  "fecha": "Martes 16 septiembre de 2025 | 23:01",
  "tags": ["sociedad", "viral", "2025"],
  "autor": "Nombre del Autor",
  "desc_autor": "Descripción del autor",
  "abstract": "Resumen de la noticia",
  "cuerpo": "Contenido completo de la noticia...",
  "multimedia": ["https://media.biobiochile.cl/..."],
  "tipo_multimedia": "imagen"
}
```

**Características**:
- ✅ Campos en español (tal como vienen del scraper)
- ✅ Fechas en formato español: "Martes 16 septiembre de 2025 | 23:01"
- ✅ Tags como array de strings
- ✅ Tipo de multimedia: "imagen", "video", "audio"
- ✅ Auto-extracción del medio desde la URL (ej: biobiochile)

## Inicio Rápido

### Prerequisitos

- Docker
- Docker Compose
- Python 3 (para script de ingesta)

### 1. Levantar el sistema

```bash
# Iniciar todos los servicios
docker-compose up -d

# Verificar que todos los servicios estén corriendo
docker-compose ps

# Ver logs en tiempo real
docker-compose logs -f
```

### Orden de Inicio

El sistema se inicia automáticamente en el orden correcto:

1. Infraestructura (postgres, elasticsearch, rabbitmq)
2. Migrator (ejecuta schema SQL)
3. ES-Init (crea índice en Elasticsearch)
4. API y Workers

Espera unos 30 segundos para que todos los servicios estén listos.

### 2. Ingestar datos desde output.json

```bash
# Ingesta directa del output.json del scraper
python3 scripts/ingest_output.py /ruta/a/output.json

# Ejemplo con el archivo de descarga
python3 scripts/ingest_output.py /home/basty/Downloads/output.json

# Con delay personalizado entre requests
python3 scripts/ingest_output.py /home/basty/Downloads/output.json --delay 0.2

# Con API remota
python3 scripts/ingest_output.py /home/basty/Downloads/output.json --api-url http://servidor:8080
```

**El script**:
- Lee el output.json directamente
- Envía cada artículo a la API sin transformaciones
- Muestra progreso en tiempo real
- Genera resumen de éxito/errores

### 3. Verificar ingesta

```bash
# Monitorear workers
docker-compose logs -f worker-indexer worker-sync

# Verificar en PostgreSQL
docker-compose exec postgres psql -U postgres -d newspress_db -c "SELECT COUNT(*) FROM news;"

# Ver artículos con tags
docker-compose exec postgres psql -U postgres -d newspress_db -c "
  SELECT n.titulo, array_agg(t.name) as tags
  FROM news n
  LEFT JOIN news_tags nt ON n.id = nt.news_id
  LEFT JOIN tags t ON nt.tag_id = t.id
  GROUP BY n.id, n.titulo
  LIMIT 5;
"

# Verificar en Elasticsearch
curl -X GET "http://localhost:9200/news/_search?pretty&size=1"

# Contar documentos
curl -X GET "http://localhost:9200/news/_count?pretty"
```

## Endpoints API

### API Ingestion

- **POST /api/v1/news** - Ingestar noticia (formato scraper)
  ```bash
  curl -X POST http://localhost:8080/api/v1/news \
    -H "Content-Type: application/json" \
    -d @artículo.json
  ```

- **GET /health** - Health check
  ```bash
  curl http://localhost:8080/health
  ```

## Acceso a Servicios

- **API Ingestion**: http://localhost:8080
- **RabbitMQ Management**: http://localhost:15672 (guest/guest)
- **Elasticsearch**: http://localhost:9200
- **PostgreSQL**: localhost:5432 (postgres/postgres123)

## Desarrollo

### Estructura del Proyecto

```
.
├── api-ingestion/          # API REST (acepta formato español)
│   ├── models/            # NewsRequest con campos en español
│   └── handlers/          # Maneja requests del scraper
├── workers/
│   ├── common/            # Configuración compartida
│   ├── indexer/           # Worker con parser de fechas españolas
│   └── sync/              # Sincronización a Elasticsearch
├── database/
│   └── migrations/        # Schema SQL
├── elasticsearch/
│   └── init/              # Mapping con soporte multimedia
├── scripts/
│   └── ingest_output.py   # Script directo para output.json
└── docker-compose.yml     # Orquestación de servicios
```

### Componentes Clave

#### API Ingestion - NewsRequest
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

#### Worker Indexer
- Parser de fechas españolas integrado
- Auto-extracción del medio desde URL
- Creación automática de tags
- Detección de duplicados por URL hash

### Logs

```bash
# Ver logs de un servicio específico
docker-compose logs -f api-ingestion
docker-compose logs -f worker-indexer
docker-compose logs -f worker-sync

# Ver logs de todos los servicios
docker-compose logs -f

# Ver logs con timestamps
docker-compose logs -f -t
```

### Detener el sistema

```bash
# Detener servicios
docker-compose down

# Detener y eliminar volúmenes (CUIDADO: elimina datos)
docker-compose down -v
```

## Características del Sistema

### Procesamiento de Fechas
El worker indexer parsea automáticamente fechas en español:
- ✅ "Martes 16 septiembre de 2025 | 23:01"
- ✅ "Viernes 20 septiembre de 2024 | 16:00"
- ✅ Detecta todos los meses en español
- ✅ Convierte a timezone de Chile (America/Santiago)

### Sistema de Tags
- Auto-creación de tags desde el array del scraper
- Relación many-to-many con noticias
- Indexados en Elasticsearch para búsqueda facetada

### Multimedia
- Soporte para múltiples URLs de multimedia
- Tipos: "imagen", "video", "audio"
- Almacenado tanto en PostgreSQL como Elasticsearch

### Detección de Duplicados
- Hash SHA256 de URL normalizada
- Previene ingesta duplicada
- Validación antes de insertar en base de datos

## Resolución de Problemas

### Error: "formato de fecha no reconocido"
Verifica que la fecha esté en el formato correcto del scraper:
```
"Día# mes de año | hora:minuto"
Ejemplo: "Martes 16 septiembre de 2025 | 23:01"
```

### Error: "failed to connect to RabbitMQ"
```bash
# Reiniciar servicios
docker-compose restart rabbitmq
docker-compose logs rabbitmq
```

### Workers no procesan mensajes
```bash
# Verificar colas en RabbitMQ
curl -u guest:guest http://localhost:15672/api/queues

# Reiniciar workers
docker-compose restart worker-indexer worker-sync
```

### PostgreSQL: "relation does not exist"
```bash
# Verificar que las migraciones se ejecutaron
docker-compose logs migrator

# Re-ejecutar migraciones
docker-compose up migrator
```

### Elasticsearch: "index not found"
```bash
# Verificar que el índice se creó
curl http://localhost:9200/_cat/indices

# Re-crear índice
docker-compose restart elasticsearch-init
```

## Consultas Útiles

### PostgreSQL

```sql
-- Ver últimas noticias ingresadas
SELECT titulo, autor, published_date, created_at
FROM news
ORDER BY created_at DESC
LIMIT 10;

-- Contar noticias por medio
SELECT ms.name, COUNT(*) as total
FROM news n
JOIN media_sources ms ON n.media_source_id = ms.id
GROUP BY ms.name;

-- Ver tags más usados
SELECT t.name, COUNT(*) as usage_count
FROM tags t
JOIN news_tags nt ON t.id = nt.tag_id
GROUP BY t.name
ORDER BY usage_count DESC
LIMIT 20;

-- Ver noticias con multimedia
SELECT n.titulo, m.media_type, COUNT(*) as media_count
FROM news n
JOIN news_multimedia m ON n.id = m.news_id
GROUP BY n.id, n.titulo, m.media_type;
```

### Elasticsearch

```bash
# Buscar por término
curl -X GET "http://localhost:9200/news/_search?pretty" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "match": {
        "content": "chile"
      }
    }
  }'

# Agregación por tags
curl -X GET "http://localhost:9200/news/_search?pretty" \
  -H 'Content-Type: application/json' \
  -d '{
    "size": 0,
    "aggs": {
      "popular_tags": {
        "terms": { "field": "tags", "size": 20 }
      }
    }
  }'

# Filtrar por tipo de multimedia
curl -X GET "http://localhost:9200/news/_search?pretty" \
  -H 'Content-Type: application/json' \
  -d '{
    "query": {
      "term": {
        "multimedia.media_type": "video"
      }
    }
  }'
```

## Próximos Pasos

- [ ] API de consultas (query API) para búsquedas
- [ ] Dashboard de monitoreo
- [ ] Tests unitarios y de integración
- [ ] Logging estructurado (JSON)
- [ ] Métricas con Prometheus
- [ ] Autenticación y autorización
- [ ] Dead letter queue para mensajes fallidos
- [ ] Retry automático con backoff exponencial
- [ ] Soporte para más idiomas/formatos de fecha
