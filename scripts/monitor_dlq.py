#!/usr/bin/env python3
"""
Script para monitorear las Dead Letter Queues (DLQ) de RabbitMQ

Este script permite:
1. Ver cuántos mensajes hay en cada DLQ
2. Inspeccionar mensajes fallidos
3. Opcionalmente, reintentar manualmente mensajes desde la DLQ

Uso:
    python3 scripts/monitor_dlq.py --status
    python3 scripts/monitor_dlq.py --inspect ingestion_queue_dlq
    python3 scripts/monitor_dlq.py --retry ingestion_queue_dlq <message_count>
"""

import argparse
import json
import sys
from datetime import datetime

import pika

# Configuración de RabbitMQ
RABBITMQ_HOST = "localhost"
RABBITMQ_PORT = 5672
RABBITMQ_USER = "guest"
RABBITMQ_PASS = "guest"

DLQ_QUEUES = ["ingestion_queue_dlq", "sync_queue_dlq"]


def get_connection():
    """Establece conexión con RabbitMQ"""
    credentials = pika.PlainCredentials(RABBITMQ_USER, RABBITMQ_PASS)
    parameters = pika.ConnectionParameters(
        host=RABBITMQ_HOST, port=RABBITMQ_PORT, credentials=credentials
    )
    return pika.BlockingConnection(parameters)


def get_queue_status(channel, queue_name):
    """Obtiene el estado de una cola"""
    try:
        method = channel.queue_declare(queue=queue_name, passive=True)
        return {
            "queue": queue_name,
            "message_count": method.method.message_count,
            "consumer_count": method.method.consumer_count,
        }
    except Exception as e:
        return {"queue": queue_name, "error": str(e)}


def show_status():
    """Muestra el estado de todas las DLQs"""
    connection = get_connection()
    channel = connection.channel()

    print("\n" + "=" * 70)
    print("📊 DEAD LETTER QUEUE STATUS")
    print("=" * 70)
    print(f"Timestamp: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n")

    for queue_name in DLQ_QUEUES:
        status = get_queue_status(channel, queue_name)

        if "error" in status:
            print(f"❌ {queue_name}: {status['error']}")
        else:
            msg_count = status["message_count"]
            icon = "✅" if msg_count == 0 else "⚠️"
            print(f"{icon} {queue_name}:")
            print(f"   Messages: {msg_count}")
            print(f"   Consumers: {status['consumer_count']}")

            if msg_count > 0:
                print(f"   ⚠️  WARNING: {msg_count} failed messages in DLQ!")
        print()

    connection.close()


def inspect_messages(queue_name, limit=10):
    """Inspecciona mensajes en una DLQ sin consumirlos"""
    connection = get_connection()
    channel = connection.channel()

    print("\n" + "=" * 70)
    print(f"🔍 INSPECTING DLQ: {queue_name}")
    print("=" * 70 + "\n")

    messages = []

    for i in range(limit):
        method, properties, body = channel.basic_get(queue=queue_name, auto_ack=False)

        if method is None:
            break

        # Rechazar el mensaje para que vuelva a la cola (no lo consumimos)
        channel.basic_nack(delivery_tag=method.delivery_tag, requeue=True)

        try:
            body_json = json.loads(body)
        except:
            body_json = body.decode("utf-8")

        message_info = {
            "index": i + 1,
            "body": body_json,
            "headers": properties.headers or {},
            "timestamp": properties.timestamp,
            "delivery_tag": method.delivery_tag,
        }
        messages.append(message_info)

    if not messages:
        print("✅ No messages in DLQ")
    else:
        for msg in messages:
            print(f"\n{'─' * 70}")
            print(f"Message #{msg['index']}")
            print(f"{'─' * 70}")

            # Headers
            if msg["headers"]:
                print("\n📋 Headers:")
                retry_count = msg["headers"].get("x-retry-count", 0)
                retry_reason = msg["headers"].get("x-retry-reason", "N/A")
                dlq_reason = msg["headers"].get("x-dlq-reason", "N/A")

                print(f"   Retry Count: {retry_count}")
                print(f"   Retry Reason: {retry_reason}")
                print(f"   DLQ Reason: {dlq_reason}")

            # Body
            print("\n📦 Body:")
            if isinstance(msg["body"], dict):
                print(f"   {json.dumps(msg['body'], indent=2, ensure_ascii=False)}")
            else:
                print(f"   {msg['body']}")

    connection.close()
    print(f"\n{'=' * 70}\n")


def retry_messages(queue_name, count, target_queue):
    """Reintenta mensajes desde la DLQ moviéndolos a la cola principal"""
    connection = get_connection()
    channel = connection.channel()

    print("\n" + "=" * 70)
    print(f"🔄 RETRYING MESSAGES FROM: {queue_name}")
    print("=" * 70 + "\n")

    retried = 0

    for i in range(count):
        method, properties, body = channel.basic_get(queue=queue_name, auto_ack=False)

        if method is None:
            print(f"⚠️  Only {retried} messages available in DLQ")
            break

        # Resetear el contador de reintentos
        new_headers = properties.headers or {}
        new_headers["x-retry-count"] = 0
        new_headers["x-manual-retry"] = True
        new_headers["x-manual-retry-timestamp"] = datetime.now().isoformat()

        # Publicar al queue original
        channel.basic_publish(
            exchange="",
            routing_key=target_queue,
            body=body,
            properties=pika.BasicProperties(
                content_type=properties.content_type,
                headers=new_headers,
                delivery_mode=2,  # persistent
            ),
        )

        # Acknowledge del mensaje en DLQ
        channel.basic_ack(delivery_tag=method.delivery_tag)
        retried += 1

        print(f"✓ Message {retried} retried to {target_queue}")

    connection.close()

    print(f"\n{'=' * 70}")
    print(f"✅ Successfully retried {retried} messages")
    print(f"{'=' * 70}\n")


def purge_dlq(queue_name):
    """Purga (elimina) todos los mensajes de una DLQ"""
    connection = get_connection()
    channel = connection.channel()

    print(f"\n⚠️  WARNING: This will permanently delete all messages from {queue_name}")
    confirm = input("Type 'YES' to confirm: ")

    if confirm == "YES":
        channel.queue_purge(queue=queue_name)
        print(f"✅ DLQ {queue_name} purged successfully\n")
    else:
        print("❌ Operation cancelled\n")

    connection.close()


def main():
    parser = argparse.ArgumentParser(
        description="Monitor and manage Dead Letter Queues"
    )
    parser.add_argument("--status", action="store_true", help="Show DLQ status")
    parser.add_argument("--inspect", metavar="QUEUE", help="Inspect messages in DLQ")
    parser.add_argument("--retry", metavar="QUEUE", help="Retry messages from DLQ")
    parser.add_argument(
        "--count", type=int, default=10, help="Number of messages to inspect/retry"
    )
    parser.add_argument("--purge", metavar="QUEUE", help="Purge all messages from DLQ")

    args = parser.parse_args()

    try:
        if args.status:
            show_status()
        elif args.inspect:
            inspect_messages(args.inspect, limit=args.count)
        elif args.retry:
            # Determinar la cola de destino (remover _dlq del nombre)
            target_queue = args.retry.replace("_dlq", "")
            retry_messages(args.retry, args.count, target_queue)
        elif args.purge:
            purge_dlq(args.purge)
        else:
            parser.print_help()

    except Exception as e:
        print(f"\n❌ Error: {e}\n")
        sys.exit(1)


if __name__ == "__main__":
    main()
