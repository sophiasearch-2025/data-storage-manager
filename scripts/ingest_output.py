#!/usr/bin/env python3
"""
Ingesta directa del output.json del scraper
El formato del scraper se envía tal cual a la API
"""

import json
import sys
import time
from typing import Any, Dict, List

import requests

def ingest_articles(
    output_file: str, api_url: str = "http://localhost:8080", delay: float = 0.1
):
    """
    Ingesta artículos directamente desde output.json

    Args:
        output_file: Ruta al archivo output.json del scraper
        api_url: URL base de la API
        delay: Delay entre requests (segundos)
    """

    # Leer output.json
    print(f"📄 Leyendo {output_file}...")
    try:
        with open(output_file, "r", encoding="utf-8") as f:
            articles = json.load(f)
    except Exception as e:
        print(f"❌ Error leyendo archivo: {e}")
        sys.exit(1)

    if not isinstance(articles, list):
        print("❌ Error: El archivo debe contener un array JSON")
        sys.exit(1)

    print(f"✅ {len(articles)} artículos encontrados\n")

    # Health check
    print(f"🏥 Verificando API en {api_url}...")
    try:
        response = requests.get(f"{api_url}/health", timeout=5)
        if response.status_code == 200:
            print("✅ API disponible\n")
        else:
            print(f"❌ API respondió con status {response.status_code}")
            sys.exit(1)
    except Exception as e:
        print(f"❌ No se puede conectar a la API: {e}")
        print("   Asegúrate de que los servicios estén corriendo: docker-compose up -d")
        sys.exit(1)

    # Ingestar artículos
    endpoint = f"{api_url}/api/v1/news"
    success = 0
    failed = 0
    errors = []

    print("🚀 Iniciando ingesta...\n")
    print("=" * 80)

    for i, article in enumerate(articles, 1):
        titulo = article.get("titulo", "Sin título")[:60]
        print(f"[{i}/{len(articles)}] {titulo}...", end=" ")

        try:
            response = requests.post(
                endpoint,
                json=article,  # Se envía tal cual, sin transformación
                headers={"Content-Type": "application/json"},
                timeout=10,
            )

            if response.status_code == 202:
                job_id = response.json().get("job_id", "N/A")
                print(f"✅ {job_id}")
                success += 1
            else:
                error_msg = response.text[:100]
                print(f"❌ {response.status_code}: {error_msg}")
                failed += 1
                errors.append(
                    {
                        "url": article.get("url", "N/A"),
                        "error": error_msg,
                        "status": response.status_code,
                    }
                )

        except Exception as e:
            print(f"❌ {str(e)[:100]}")
            failed += 1
            errors.append({"url": article.get("url", "N/A"), "error": str(e)})

        # Delay para no saturar
        if delay > 0 and i < len(articles):
            time.sleep(delay)

    # Resumen
    print("=" * 80)
    print("\n📊 RESUMEN DE INGESTA")
    print("=" * 80)
    print(f"Total:    {len(articles)}")
    print(f"Exitosos: {success} ✅")
    print(f"Fallidos: {failed} ❌")
    print(f"Tasa de éxito: {success / len(articles) * 100:.1f}%")

    if errors:
        print(f"\n⚠️  Primeros errores:")
        for error in errors[:5]:
            print(f"  • {error['url']}: {error.get('error', 'Unknown')}")

    print("\n💡 Monitorear procesamiento:")
    print("   docker-compose logs -f worker-indexer worker-sync")

    return success, failed


if __name__ == "__main__":
    import argparse

    parser = argparse.ArgumentParser(
        description="Ingesta directa de output.json del scraper"
    )
    parser.add_argument("output_file", help="Ruta al archivo output.json")
    parser.add_argument(
        "--api-url",
        default="http://localhost:8080",
        help="URL de la API (default: http://localhost:8080)",
    )
    parser.add_argument(
        "--delay",
        type=float,
        default=0.1,
        help="Delay entre requests en segundos (default: 0.1)",
    )
                       help='Delay entre requests en segundos (default: 0.1)')

    args = parser.parse_args()

    success, failed = ingest_articles(args.output_file, args.api_url, args.delay)

    sys.exit(0 if failed == 0 else 1)
