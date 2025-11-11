## 🚀 NewsPress - Data Storage Manager

Sistema de gestión de almacenamiento de noticias con arquitectura de microservicios.

## 🌟 Visión General

Este proyecto implementa una solución de almacenamiento y búsqueda para datos de noticias utilizando una arquitectura basada en colas de mensajes (microservicios).

Para una visión completa, consulta los siguientes documentos:

| Documento | Descripción |
| :--- | :--- |
| **[Arquitectura](arquitectura.md)** | Detalle de los componentes y el flujo de datos. |
| **[Decisiones Arquitectónicas](decisiones.md)** | Justificación de la elección de tecnologías (PostgreSQL, Elasticsearch, RabbitMQ). |
| **[Despliegue y Operación](deploy.md)** | Instrucciones de inicio rápido, endpoints y manejo de logs. |
| **[Requisitos Técnicos](requisitos.md)** | Prerrequisitos de software y configuración de desarrollo. |

---

## 💻 Desarrollo

Consulta la [sección de Desarrollo](deploy.md#desarrollo) para la estructura del proyecto y comandos de logs.

### Acceso a Servicios

| Servicio | Acceso | Credenciales (si aplica) |
| :--- | :--- | :--- |
| **API Ingestion** | http://localhost:8080 | |
| **RabbitMQ Management** | http://localhost:15672 | guest / guest |
| **Elasticsearch** | http://localhost:9200 | |
| **PostgreSQL** | localhost:5432 | postgres / postgres123 |

## 🚧 Pendientes

* [ ] API de consultas (query API)
* [ ] Tests unitarios y de integración
* [ ] Logging estructurado
* [ ] Métricas y monitoreo
* [ ] Autenticación y autorización
* [ ] Dead letter queue para mensajes fallidos