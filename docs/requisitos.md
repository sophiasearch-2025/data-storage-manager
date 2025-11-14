# 📎 Requisitos Técnicos y Funcionales

Este documento consolida los requisitos de software necesarios para la operación del sistema y define el formato de datos que el microservicio **API Ingestion** está diseñado para recibir.

## 1. ⚙️ Prerrequisitos de Software

Para el correcto despliegue local y el desarrollo del proyecto, se requiere la instalación de:

* **Docker:** Necesario para ejecutar el motor de contenedores y los microservicios.
* **Docker Compose (Plugin):** Esencial para la orquestación y gestión de los múltiples servicios definidos en el archivo `docker-compose.yml`.

---

## 2. ⚡ Requisitos Funcionales

### 2.1. Formato de Entrada de Noticias (Scraper Nativo)

El microservicio está diseñado para recibir y procesar directamente el formato de datos generado por el *scraper* (en español). El **Worker Indexer** se encarga internamente de la traducción de campos y el *parsing* de fechas.

**Estructura JSON Esperada:**

```json
{
  "url": "[https://www.biobiochile.cl/noticias/](https://www.biobiochile.cl/noticias/)...",
  "titulo": "Título de la noticia",
  "fecha": "Martes 16 septiembre de 2025 | 23:01",
  "tags": ["sociedad", "viral", "2025"],
  "autor": "Nombre del Autor",
  "desc_autor": "Descripción del autor",
  "abstract": "Resumen de la noticia",
  "cuerpo": "Contenido completo de la noticia...",
  "multimedia": ["[https://media.biobiochile.cl/](https://media.biobiochile.cl/)..."],
  "tipo_multimedia": "imagen"
}
```

### 2.2 ⚡ Requisitos Funcionales (Endpoints)

El microservicio **API Ingestion** expone los siguientes endpoints, que definen los requisitos funcionales de interacción con el cliente:

### API Ingestion (http://localhost:8080)

| Método | Ruta | Propósito |
| :--- | :--- | :--- |
| `GET` | `/health` | Realiza una verificación de estado (`Health Check`) para confirmar que el servicio está activo. |
| `POST` | `/api/v1/news` | **Ingesta de Noticias.** Recibe una nueva noticia en formato JSON y la encola para su procesamiento asíncrono. |