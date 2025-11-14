# 🚀 Despliegue y Operación

Este sistema utiliza **Docker Compose** para orquestar la infraestructura de microservicios. Todas las operaciones de despliegue se ejecutan desde la **raíz del repositorio**.

## 1. ⚙️ Prerrequisitos de Entorno

Asegúrate de tener instalados y configurados los siguientes requisitos:

* **Docker:** El motor principal debe estar instalado y el servicio del demonio debe estar corriendo.
* **Docker Compose (Plugin):** Necesario para la orquestación.

---

## 2. 🚀 Levantar el Sistema (Despliegue Local)

Ejecuta estos comandos en orden para poner el sistema en línea:

1.  **Configuración de Variables de Entorno:**
    Copia el archivo de ejemplo para crear tu configuración local. Este archivo (`.env`) es vital para puertos, credenciales y URLs de RabbitMQ.
    ```bash
    cp .env.example .env
    ```

2.  **Iniciar Todos los Servicios:**
    El comando construye (si es necesario) e inicia todos los contenedores en modo *detached* (segundo plano).
    ```bash
    docker compose up -d
    ```

3.  **Verificación Rápida:**
    Confirma que todos los contenedores han iniciado correctamente.
    ```bash
    docker compose ps
    ```
    > **Resultado Esperado:** Todos los servicios (APIs, Workers, PostgreSQL, etc.) deben mostrar el estado **`Up`**.

### Flujo y Orden de Inicio (Anticorrupción)

El `docker-compose.yml` asegura un orden de arranque estricto para evitar fallos de conexión:

1.  **Infraestructura:** (Postgres, Elasticsearch, RabbitMQ) se inician primero.
2.  **Inicialización:** `Migrator` (ejecuta el esquema SQL) y **`ES-Init`** (crea el índice `news` en Elasticsearch) se ejecutan **una sola vez**.
3.  **Aplicaciones:** `API` y `Workers` se inician solo después de que la infraestructura y la inicialización hayan finalizado.

---

## 3. 🧪 Pruebas Funcionales (Post-Deploy)

Una vez que el sistema está `Up`, verifica el flujo completo (API → RabbitMQ → PostgreSQL → Elasticsearch).

1.  **Ingestar Noticia:** Envía una noticia de prueba a la API. El sistema debe responder con `202 Accepted` ya que el procesamiento es asíncrono.
    ```bash
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
    ```

2.  **Verificar en PostgreSQL:** Comprueba si el `Worker Indexer` guardó la noticia en la tabla `news`.
    ```bash
    docker exec data-storage-manager-postgres psql -U postgres -d newspress -c "SELECT id, title FROM news;"
    ```

3.  **Verificar en Elasticsearch:** Confirma que el `Worker Sync` sincronizó la noticia al motor de búsqueda.
    ```bash
    curl -X GET "http://localhost:9200/news/_search?pretty"
    ```

---

## 4. 🛑 Detener y Limpiar

* **Detener servicios (manteniendo datos):** Detiene los contenedores, pero **conserva los volúmenes de datos**. Es el comando más seguro para detener el desarrollo diario.
    ```bash
    docker compose down
    ```

* **Detener y eliminar volúmenes (PELIGRO: elimina datos permanentes):** Elimina contenedores, redes **y los volúmenes de datos** persistentes (PostgreSQL, Elasticsearch). Úsalo solo cuando desees reiniciar el proyecto desde cero.
    ```bash
    docker compose down -v
    ```

---

## 5. 🔍 Logs y Troubleshooting

Para monitorear el estado en tiempo real y depurar problemas:

* **Ver logs de todos los servicios en tiempo real:**
    ```bash
    docker compose logs -f
    ```

* **Ver logs de un servicio específico:** Útil para aislar errores. Reemplaza `api-ingestion` por el nombre del servicio a inspeccionar.
    ```bash
    docker compose logs -f api-ingestion
    ```