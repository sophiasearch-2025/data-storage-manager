# Instrucciones de ejecución
1. Crear la base de datos en PostgreSQL
```bash
CREATE DATABASE nombre_de_tu_base;
```

2️. Ejecutar el script SQL para crear las tablas
```bash
psql -U tu_usuario -d nombre_de_tu_base -f crear_tablas.sql
```

3️. Configurar el archivo .env

Editar el archivo .env con tus datos de conexión:
```bash
DB_HOST=localhost
DB_PORT=5432
DB_NAME=nombre_de_tu_base
DB_USER=tu_usuario
DB_PASSWORD=tu_contraseña
```

4️. Instalar dependencias
```bash
pip install -r requirements.txt
```

5️. Ejecutar el script para cargar los datos
```bash
python cargar_dataset.py
```

6️. (Opcional) Verificar la carga en PostgreSQL
```bash
SELECT COUNT(*) FROM noticia;
SELECT COUNT(*) FROM medios_prensa;

SELECT COUNT(*) FROM noticia_detalle;
```

