## 5. 📎 requisitos.md

Este archivo contiene una sección para los requisitos y la lista de endpoints.

```markdown
# 📎 Requisitos Técnicos y Endpoints

## Prerrequisitos de Desarrollo

Para trabajar en el código o desplegar la infraestructura, necesitarás:

* **Docker**
* **Docker Compose** (Plugin)

## Endpoints Disponibles

### API Ingestion

La API de Ingesta se expone en `http://localhost:8080` y acepta los siguientes endpoints:

| Método | Ruta | Descripción |
| :--- | :--- | :--- |
| `GET` | `/health` | Health check del servicio. |
| `POST` | `/api/v1/news` | Ingestar una nueva noticia (manda el mensaje a RabbitMQ). |