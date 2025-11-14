# 📎 Requisitos Técnicos y Funcionales

Este documento lista el software necesario para el desarrollo y despliegue local, así como los requisitos funcionales expuestos por el microservicio de ingesta.

## 1. ⚙️ Prerrequisitos de Software

Para el correcto despliegue local y el desarrollo del proyecto, se requiere la instalación de:

* **Docker:** Necesario para ejecutar el motor de contenedores y los microservicios.
* **Docker Compose (Plugin):** Esencial para la orquestación y gestión de los múltiples servicios definidos en el archivo `docker-compose.yml`.

---

## 2. ⚡ Requisitos Funcionales (Endpoints)

El microservicio **API Ingestion** expone los siguientes endpoints, que definen los requisitos funcionales de interacción con el cliente:

### API Ingestion (http://localhost:8080)

| Método | Ruta | Propósito |
| :--- | :--- | :--- |
| `GET` | `/health` | Realiza una verificación de estado (`Health Check`) para confirmar que el servicio está activo. |
| `POST` | `/api/v1/news` | **Ingesta de Noticias.** Recibe una nueva noticia en formato JSON y la encola para su procesamiento asíncrono. |