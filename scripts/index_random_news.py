from elasticsearch import Elasticsearch

# Conectar a ElasticSearch en localhost
es = Elasticsearch(hosts=["http://localhost:9200"])

# Nombre del índice por defecto
INDEX_NAME = "noticias"

# Crear índice si no existe
if not es.indices.exists(index=INDEX_NAME):
    es.indices.create(index=INDEX_NAME)
    print(f"Índice '{INDEX_NAME}' creado")
else:
    print(f"Índice '{INDEX_NAME}' ya existe")

# Lista de noticias de prueba
noticias = [
    {
        "date": "Oct 28, 2025",
        "country": "Chile",
        "media_outlet": "El Sur",
        "title": "Avance educativo para jóvenes",
        "text": "Se impulsa un nuevo programa educativo para la reinserción juvenil.",
        "url": "https://www.elsur.cl/noticia1"
    },
    {
        "date": "Oct 28, 2025",
        "country": "Chile",
        "media_outlet": "La Tercera",
        "title": "Innovación tecnológica en escuelas",
        "text": "Nuevas herramientas digitales se implementan en aulas chilenas.",
        "url": "https://www.latercera.cl/noticia2"
    },
    {
        "date": "Oct 28, 2025",
        "country": "Chile",
        "media_outlet": "El Mercurio",
        "title": "Cultura y deporte en colegios",
        "text": "Programas deportivos y culturales para estudiantes de todo el país.",
        "url": "https://www.elmercurio.cl/noticia3"
    }
]

# Indexar cada noticia
for noticia in noticias:
    es.index(index=INDEX_NAME, document=noticia)

print("Noticias de prueba indexadas correctamente.")