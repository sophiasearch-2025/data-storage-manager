# ElasticSearch en data-storage-manager

Este repositorio contiene la configuración para levantar ElasticSearch usando Docker.

## Requisitos

* Docker >= 20
* Docker Compose >= 1.29

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

3. Verificar que está corriendo:

```bash
docker ps
# o
docker-compose ps
```

4. Probar ElasticSearch (si tienes puerto expuesto):

```bash
curl http://localhost:9200
```

5. Para detener el servicio:

```bash
docker-compose down
```

## Notas

* Los datos se guardan en el volumen `es_data`.
* Para desarrollo local, puedes exponer el puerto 9200 agregando temporalmente en docker-compose:

```yaml
ports:
  - "9200:9200"
```
