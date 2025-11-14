from elasticsearch import Elasticsearch

# Cliente (con compatibilidad para ES 8 si usas la v9 del client)
es = Elasticsearch(
    "http://localhost:9200",
    headers={"Accept": "application/vnd.elasticsearch+json; compatible-with=8"},
)

INDEX = "noticias"


def buscar_texto_libre(texto, size=10):
    """
    Búsqueda tipo Google en titulo, texto_noticia y medio.nombre
    """
    query = {
        "query": {
            "multi_match": {
                "query": texto,
                "fields": ["titulo", "texto_noticia", "medio.nombre"],
                "fuzziness": "AUTO",  # tolera errores de tipeo
            }
        }
    }

    resp = es.search(index=INDEX, body=query, size=size)
    print(f"\n Resultados para: {texto}")
    for hit in resp["hits"]["hits"]:
        src = hit["_source"]
        print(f"- ({hit['_score']:.2f}) {src.get('titulo')} [{src.get('medio', {}).get('nombre')}]")
    print(f"Total (aprox.): {resp['hits']['total']}")


def buscar_por_pais(pais, size=10):
    """
    Buscar noticias por país del medio
    """
    query = {
        "query": {
            "match": {
                "medio.pais": pais
            }
        }
    }

    resp = es.search(index=INDEX, body=query, size=size)
    print(f"\n Noticias cuyo medio es de: {pais}")
    for hit in resp["hits"]["hits"]:
        src = hit["_source"]
        print(f"- {src.get('titulo')} ({src.get('medio', {}).get('nombre')})")


def buscar_por_rango_fechas(desde, hasta=None, size=10):
    """
    Buscar noticias desde una fecha (y opcionalmente hasta otra)
    Formato fecha: 'YYYY-MM-DD'
    """
    range_filter = {"gte": desde}
    if hasta:
        range_filter["lte"] = hasta

    query = {
        "query": {
            "range": {
                "fecha_subida": range_filter
            }
        },
        "sort": [
            {"fecha_subida": "desc"}
        ]
    }

    resp = es.search(index=INDEX, body=query, size=size)
    print(f"\n Noticias desde {desde}" + (f" hasta {hasta}" if hasta else ""))
    for hit in resp["hits"]["hits"]:
        src = hit["_source"]
        print(f"- {src.get('fecha_subida')} | {src.get('titulo')}")


def buscar_por_texto_y_tag(texto, tag, size=10):
    """
    Buscar combinando texto libre + un tag exacto
    (cuando tengas tags reales en el índice)
    """
    query = {
        "query": {
            "bool": {
                "must": [
                    {
                        "multi_match": {
                            "query": texto,
                            "fields": ["titulo", "texto_noticia"],
                            "fuzziness": "AUTO",
                        }
                    }
                ],
                "filter": [
                    {"term": {"tags": tag}}
                ]
            }
        }
    }

    resp = es.search(index=INDEX, body=query, size=size)
    print(f"\n Noticias sobre '{texto}' con tag '{tag}'")
    for hit in resp["hits"]["hits"]:
        src = hit["_source"]
        print(f"- {src.get('titulo')} | tags={src.get('tags')}")


if __name__ == "__main__":
    # Comprueba que conecta
    print(" Conectado a ES:", es.info()["version"]["number"])

    # Ejemplos de uso:
    buscar_texto_libre("incendio")
    buscar_texto_libre("Valdivia")
    buscar_por_pais("chile")
    buscar_por_rango_fechas("2025-01-01")
    
