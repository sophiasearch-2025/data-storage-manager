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

## Notas

* Los datos de ElasticSearch se guardan en el volumen `es_data`.
* Mantener el puerto 9200 expuesto para que scripts Python se conecten al contenedor.
