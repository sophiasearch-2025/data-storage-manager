# ⚖️ Decisiones Arquitectónicas Clave

Este proyecto utiliza un patrón de **CQRS simplificado** (Command Query Responsibility Segregation) donde la base de datos relacional maneja las escrituras de datos maestros, y el motor de búsqueda maneja las consultas.

## 1. PostgreSQL (SQL) como Almacenamiento Primario

* **Decisión:** Usar PostgreSQL para el almacenamiento persistente.
* **Justificación:** Se utiliza como la fuente **única y confiable de verdad** (Single Source of Truth). Garantiza la **integridad transaccional** (ACID) y la consistencia de los datos antes de cualquier indexación.

## 2. Elasticsearch para Consultas y Búsqueda

* **Decisión:** Usar Elasticsearch en un flujo asíncrono.
* **Justificación:** Es un motor de búsqueda de texto completo altamente optimizado para consultas rápidas, filtros complejos y búsquedas por relevancia, tareas para las que PostgreSQL no es ideal. Mejora significativamente el rendimiento de la **API de Consultas** (pendiente de implementar).

## 3. RabbitMQ (Colas de Mensajes) para Asincronía

* **Decisión:** Implementar RabbitMQ para desacoplar la ingesta y el procesamiento.
* **Justificación:**
    * **Resiliencia:** Si PostgreSQL o Elasticsearch están caídos, la API de Ingesta puede seguir recibiendo tráfico y los mensajes se almacenan en la cola, evitando la pérdida de datos y fallos en cascada.
    * **Escalabilidad:** Permite escalar los workers (`Indexer` y `Sync`) de forma independiente de la API de Ingesta, dependiendo de la carga de procesamiento necesaria.