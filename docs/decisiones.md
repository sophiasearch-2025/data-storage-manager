# ⚖️ Decisiones Arquitectónicas y Funcionales

Este documento justifica las decisiones clave de diseño tomadas para el **Data Storage Manager**, desde la elección de las tecnologías de infraestructura hasta la implementación de lógica de negocio específica.

## 1. 🌐 Decisiones de Arquitectura de Infraestructura

### A. Adopción de PostgreSQL como Fuente de Verdad (SoT)

* **Decisión:** Utilizar PostgreSQL como la base de datos relacional primaria.
* **Justificación:** Garantiza la **integridad transaccional (ACID)** y la **consistencia** de los datos. Es la fuente única y confiable de verdad (Single Source of Truth) antes de cualquier indexación secundaria.

### B. Uso de Elasticsearch para Consultas

* **Decisión:** Implementar Elasticsearch de forma asíncrona para la capa de búsqueda.
* **Justificación:** ES es un motor de búsqueda de texto completo optimizado que es superior a PostgreSQL para realizar búsquedas complejas por relevancia, filtrado y analíticas de datos, mejorando el rendimiento de la futura API de consultas.

### C. Implementación de RabbitMQ para Asincronía

* **Decisión:** Utilizar RabbitMQ como broker de mensajes para desacoplar los microservicios.
* **Justificación:** Proporciona **resiliencia** y **escalabilidad**. Si un servicio (como PostgreSQL o Elasticsearch) está inactivo, los mensajes persisten en la cola, evitando la pérdida de datos y permitiendo escalar los workers de forma independiente a la API de ingesta.

---

## 2. 🧩 Decisiones de Lógica de Negocio (Worker Indexer)

El **Worker Indexer** integra lógica compleja para manejar las particularidades del formato de entrada, lo cual fue una decisión clave para **simplificar el trabajo del scraper** y centralizar la inteligencia de procesamiento.

### A. Procesamiento de Fechas en Español

* **Decisión:** Integrar un parser de fechas que entienda el formato español nativo del scraper.
* **Justificación:** Simplifica el *scraper* al permitirle enviar la fecha en su formato nativo (`"Martes 16 septiembre de 2025 | 23:01"`) sin necesidad de transformación previa. El *worker* se encarga de convertir y normalizar estas fechas a un formato universal (`TIMESTAMP WITH TIME ZONE`) y a la zona horaria de Chile (`America/Santiago`).

### B. Detección de Duplicados

* **Decisión:** Implementar detección de duplicados por medio de un hash de la URL normalizada.
* **Justificación:** Es un requisito fundamental para mantener la **calidad del dataset** y la **integridad de la base de datos**. Al generar un hash SHA256 de la URL normalizada, se previene la ingesta de noticias ya procesadas, validando el dato antes de insertarlo en la base de datos.

### C. Sistema de Tags Automático

* **Decisión:** Permitir al *worker* crear y gestionar tags automáticamente.
* **Justificación:** Facilita la clasificación y búsqueda posterior. El sistema maneja una relación *many-to-many* entre noticias y tags, y los indexa en Elasticsearch para permitir la búsqueda facetada (filtrado por múltiples tags).

### D. Soporte Explícito para Multimedia

* **Decisión:** Almacenar la URL y el tipo de multimedia.
* **Justificación:** Permite consultar y filtrar la base de datos y Elasticsearch por el tipo de contenido (`imagen`, `video`, `audio`), lo cual es un requisito clave para el desarrollo de la futura API de consultas.