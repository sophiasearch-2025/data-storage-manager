# 🚀 Despliegue y Operación

Este sistema está diseñado para ser desplegado fácilmente utilizando Docker Compose.

## 🛠️ Inicio Rápido

### Prerequisitos

Estos requisitos son necesarios para levantar el sistema de forma local:

* **Docker**
* **Docker Compose** (Plugin)

### Levantar el sistema

1.  **Clonar variables de entorno:** Copia el archivo de ejemplo para crear tu configuración local.
    ```bash
    cp .env.example .env
    ```

2.  **Iniciar todos los servicios:**
    ```bash
    docker compose up -d
    ```

3.  **Verificar que todos los servicios estén corriendo:**
    ```bash
    docker compose ps
    ```

### Orden de Inicio

El `docker-compose.yml` maneja automáticamente la dependencia de los servicios, asegurando el orden correcto:

1.  Infraestructura (Postgres, Elasticsearch, RabbitMQ).
2.  Migrator (ejecuta el esquema SQL inicial).
3.  ES-Init (crea el índice `news` en Elasticsearch).
4.  API y Workers (inician el procesamiento).

---

## 🧪 Probar el Sistema (Flujo Completo)

Ejecuta el siguiente comando `curl` para enviar una noticia de prueba a la API de Ingesta y verifica el flujo:

```bash
# Enviar una noticia de prueba
curl -X POST http://localhost:8080/api/v1/news \
  -H "Content-Type: application/json" \
  -d '{
    "url": "[https://example.com/noticia1](https://example.com/noticia1)",
    "title": "Noticia de prueba",
    "content": "Este es el contenido de la noticia de prueba",
    "abstract": "Resumen de la noticia",
    "author": "Autor de Prueba",
    "media_outlet": "El Medio",
    "country": "chile",
    "published_date": "2025-11-07T10:00:00Z"
  }'

# Verificar que se guardó en PostgreSQL
docker exec data-storage-manager-postgres psql -U postgres -d newspress -c "SELECT id, title FROM news;"

# Verificar que se sincronizó en Elasticsearch
curl -X GET "http://localhost:9200/news/_search?pretty"