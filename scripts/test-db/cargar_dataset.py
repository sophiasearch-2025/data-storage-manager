import os
import psycopg2
import pandas as pd
from urllib.parse import urlparse
from dotenv import load_dotenv  # 👈 NUEVO

# 🔹 Cargar variables del archivo .env
load_dotenv()

# 🔹 Leer variables de entorno
DB_NAME = os.getenv("DB_NAME")
DB_USER = os.getenv("DB_USER")
DB_PASS = os.getenv("DB_PASS")
DB_HOST = os.getenv("DB_HOST")
DB_PORT = os.getenv("DB_PORT", "5432")

df = pd.read_csv("dataset_prueba.csv")

try:
    print(f"Conectando a la base {DB_NAME} en {DB_HOST}:{DB_PORT} como {DB_USER}...")  # opcional
    conn = psycopg2.connect(
        dbname=DB_NAME, user=DB_USER, password=DB_PASS,
        host=DB_HOST, port=DB_PORT
    )
    cur = conn.cursor()

    medios_insertados = 0
    noticias_insertadas = 0

    medios_unicos = df["media_outlet"].unique()
    medio_id_map = {}

    for medio in medios_unicos:
        subset = df[df["media_outlet"] == medio]
        pais = subset["country"].iloc[0]
        ejemplo_url = subset["url"].iloc[0]
        dominio = urlparse(ejemplo_url).scheme + "://" + urlparse(ejemplo_url).netloc

        cur.execute("SELECT id_medio FROM medios_prensa WHERE nombre = %s;", (medio,))
        existe = cur.fetchone()
        if existe:
            medio_id_map[medio] = existe[0]
            continue

        cur.execute("""
            INSERT INTO medios_prensa (nombre, pais, url_main_page, descripcion)
            VALUES (%s, %s, %s, %s)
            RETURNING id_medio;
        """, (medio, pais, dominio, f"Medio de comunicación: {medio}"))
        medio_id_map[medio] = cur.fetchone()[0]
        medios_insertados += 1

    for _, row in df.iterrows():
        fecha = pd.to_datetime(row["date"], format="%b %d, %Y @ %H:%M:%S.%f", errors="coerce").date()
        titulo = str(row["title"])[:500]
        url = row["url"]
        medio = row["media_outlet"]
        texto = row["text"]
        id_medio = medio_id_map[medio]

        cur.execute("SELECT id_noticia FROM noticia WHERE titulo = %s AND id_medio = %s;", (titulo, id_medio))
        if cur.fetchone():
            continue

        cur.execute("""
            INSERT INTO noticia (titulo, url, fecha_subida, id_medio)
            VALUES (%s, %s, %s, %s)
            RETURNING id_noticia;
        """, (titulo, url, fecha, id_medio))
        id_noticia = cur.fetchone()[0]
        cur.execute("INSERT INTO noticia_detalle (texto_noticia, id_noticia) VALUES (%s, %s);", (texto, id_noticia))
        noticias_insertadas += 1

    conn.commit()
    print(f"✅ Carga completada.\nMedios nuevos: {medios_insertados}\nNoticias nuevas: {noticias_insertadas}")

except Exception as e:
    print("❌ Error:", e)
    conn.rollback()
finally:
    if conn:
        cur.close()
        conn.close()
