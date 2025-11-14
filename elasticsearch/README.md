# ElasticSearch en data-storage-manager

Este repositorio contiene la configuración para levantar ElasticSearch usando Docker y cargar datos de prueba.

## Requisitos

* Docker >= 20
* Docker Compose >= 1.29
* Python 3 + pip (para ejecutar scripts de indexación)

## Instalación de Docker y Docker Compose (Linux / Ubuntu)

1. Actualizar sistema y paquetes:

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install curl wget git -y
```

2. Instalar Docker y Docker Compose:

```bash
sudo apt install docker.io docker-compose -y
sudo systemctl enable --now docker
```

3. Verificar instalación:

```bash
docker --version
docker-compose --version
```

## Levantar ElasticSearch localmente

1. Entrar a la carpeta `elasticsearch`:

```bash
cd elasticsearch
```

2. Levantar el contenedor:

```bash
docker-compose up -d
```

* El puerto 9200 está expuesto para poder conectarse desde scripts Python (`localhost:9200`).

3. Verificar que está corriendo:

```bash
docker ps
# o
docker-compose ps
```

4. Probar ElasticSearch:

```bash
curl http://localhost:9200
```

5. Para detener el servicio:

```bash
docker-compose down
```

## Indexar datos en ElasticSearch

1. Entrar a la carpeta de scripts:

```bash
cd ../scripts
```

2. Crear un entorno virtual (recomendado):

```bash
python3 -m venv venv
source venv/bin/activate
```

3. Instalar dependencias:

```bash
pip install -r requirements.txt
```

4. Ejecutar script de indexación:

```bash
python index_random_news.py
```

* Esto indexa datos de prueba en el índice `noticias`.
* Para indexar CSV reales, usar el script correspondiente (`index_data.py`) siguiendo la misma lógica.

5. Salir del entorno virtual:

```bash
deactivate
```
## Buscador de noticias con Elasticsearch y Python
* Este script (`buscar_noticias.py`) permite realizar distintas **búsquedas sobre un índice de noticias en Elasticsearch** usando Python.

1. Ejecutar script de busqueda:

```bash
python3 buscar_noticias.py
```
* El script ejecuta algunas búsquedas de ejemplo y muestra los resultados por consola.

## Estructura básica de las búsquedas (queries)
Todas las búsquedas se basan en la API search de Elasticsearch:
```bash
resp = es.search(index=INDEX, body=query, size=size)
```
* index: el índice donde se buscan datos (`noticias`).

* body: un diccionario query con la consulta en formato Elasticsearch.

* size: máximo número de resultados a devolver.

## Funciones de búsqueda incluidas (en proceso)
* 1.- `buscar_texto_libre(texto, size=10)`

Búsqueda tipo Google sobre varios campos.

Qué hace `multi_match`:
Busca el texto en varios campos a la vez (`titulo`, `texto_noticia`, `medio.nombre`).

Qué hace `fuzziness: "AUTO"` permite pequeños errores ortográficos (búsqueda difusa).

Uso de ejemplo:
```bash
buscar_texto_libre("incendio")
buscar_texto_libre("Valdivia")
```

* 2.- `buscar_por_pais(pais, size=10)`

Filtra noticias según el país del medio.

Qué hace `match`:
Busca documentos donde `medio.pais` contenga el término dado.

Uso de ejemplo:
```bash
buscar_por_pais("Chile")
```

* 3.- `buscar_por_rango_fechas(desde, hasta=None, size=10)`

Permite buscar noticias en un rango de fechas usando el campo fecha_subida.

Formato fecha: 'YYYY-MM-DD'

Qué hace range: 

`gte`: mayor o igual a la fecha `desde`.

`lte`: menor o igual a la fecha `hasta` (opcional).

Uso de ejemplo:
```bash
buscar_por_rango_fechas("2025-01-01")
buscar_por_rango_fechas("2025-01-01", "2025-12-31")
```
## Notas

* Los datos de ElasticSearch se guardan en el volumen `es_data`.
* Mantener el puerto 9200 expuesto para que scripts Python se conecten al contenedor.
