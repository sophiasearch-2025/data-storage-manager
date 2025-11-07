#!/bin/sh

# Wait for Elasticsearch to be ready
echo "Waiting for Elasticsearch to be ready..."
until curl -s http://elasticsearch:9200/_cluster/health | grep -q '"status":"green"\|"status":"yellow"'; do
  echo "Elasticsearch is unavailable - sleeping"
  sleep 5
done

echo "Elasticsearch is ready!"

# Check if index already exists
if curl -s -o /dev/null -w "%{http_code}" http://elasticsearch:9200/news | grep -q "200"; then
  echo "Index 'news' already exists, skipping creation"
  exit 0
fi

# Create the news index with proper mapping
echo "Creating 'news' index with Spanish analyzer..."

curl -X PUT "http://elasticsearch:9200/news" -H 'Content-Type: application/json' -d'
{
  "settings": {
    "number_of_shards": 1,
    "number_of_replicas": 0,
    "analysis": {
      "analyzer": {
        "spanish_analyzer": {
          "type": "spanish"
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "news_id": {
        "type": "keyword"
      },
      "title": {
        "type": "text",
        "analyzer": "spanish_analyzer",
        "fields": {
          "keyword": {
            "type": "keyword"
          }
        }
      },
      "content": {
        "type": "text",
        "analyzer": "spanish_analyzer"
      },
      "abstract": {
        "type": "text",
        "analyzer": "spanish_analyzer"
      },
      "author": {
        "type": "keyword"
      },
      "author_description": {
        "type": "text",
        "analyzer": "spanish_analyzer"
      },
      "media_source": {
        "properties": {
          "id": {
            "type": "keyword"
          },
          "name": {
            "type": "keyword"
          },
          "country": {
            "type": "keyword"
          }
        }
      },
      "published_date": {
        "type": "date"
      },
      "url": {
        "type": "keyword"
      },
      "multimedia": {
        "type": "keyword"
      },
      "tags": {
        "type": "keyword"
      },
      "status": {
        "type": "keyword"
      },
      "indexed_at": {
        "type": "date"
      }
    }
  }
}'

echo ""
echo "Index 'news' created successfully!"
