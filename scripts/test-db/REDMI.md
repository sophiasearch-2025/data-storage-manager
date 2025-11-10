🧭 Instrucciones de ejecución

1️⃣ Crear la base de datos en PostgreSQL

CREATE DATABASE nombre_de_tu_base;


2️⃣ Ejecutar el script SQL para crear las tablas

psql -U tu_usuario -d nombre_de_tu_base -f crear_tablas.sql


3️⃣ Configurar el archivo .env

Editar el archivo .env con tus datos de conexión:

DB_HOST=localhost
DB_PORT=5432
DB_NAME=nombre_de_tu_base
DB_USER=tu_usuario
DB_PASSWORD=tu_contraseña


4️⃣ Instalar dependencias

pip install -r requirements.txt


5️⃣ Ejecutar el script para cargar los datos

python cargar_dataset.py


6️⃣ (Opcional) Verificar la carga en PostgreSQL

SELECT COUNT(*) FROM noticia;
SELECT COUNT(*) FROM medios_prensa;
SELECT COUNT(*) FROM noticia_detalle;