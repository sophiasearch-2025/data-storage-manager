# 🚀 Data-Storage-Manager

> Este subsistema se encarga de gestionar la base de datos de Sophia Search con la información recopilada por el subsistema de Recopilación de Datos


## 📋 Tabla de Contenidos

* [Interaccion con otros subsistemas](#interaccion-subsistemas)
* [Documentacion Interna] (#documentacion-interna)
* [Estado Subsistema] (#estado-subsistema)
---

## 💡 Interaccion con otros subsistemas

* Recopilacion de datos
* Consulta y analisis

## ✨ Documentacion interna

Enlace a los documentos principales del subsistema:

- [Arquitectura] (./arquitectura.md)
- [Decisiones-Tecnicas] (./decisiones.md)
- [Requisitos] (./requisitos.md)
- [Despliegue] (./deploy.md)
- [Diagramas] (./diagramas
)   

### Componentes

- **PostgreSQL**: Base de datos relacional para almacenamiento persistente
- **Elasticsearch**: Motor de búsqueda para consultas rápidas
- **RabbitMQ**: Cola de mensajes para procesamiento asíncrono
- **API Ingestion**: API REST para recibir noticias
- **Worker Indexer**: Procesa noticias y las guarda en PostgreSQL
- **Worker Sync**: Sincroniza noticias de PostgreSQL a Elasticsearch

## 🛠️ Estado del subsistema
- Listo para pruebas
