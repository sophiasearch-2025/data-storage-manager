CREATE TABLE media_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    country VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabla: tags (Etiquetas/Categorías)
CREATE TABLE tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabla: news (Noticias)
CREATE TABLE news (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(500) NOT NULL,
    content TEXT,
    abstract TEXT,
    author VARCHAR(255),
    author_description TEXT,
    media_source_id UUID REFERENCES media_sources(id),
    published_date TIMESTAMP NOT NULL,
    url VARCHAR(1000) NOT NULL,
    url_hash VARCHAR(64) UNIQUE NOT NULL,
    content_hash VARCHAR(64) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Tabla: news_multimedia (URLs de imágenes/videos)
CREATE TABLE news_multimedia (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    news_id UUID REFERENCES news(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    media_type VARCHAR(50) DEFAULT 'image',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabla: news_tags (Relación muchos a muchos)
CREATE TABLE news_tags (
    news_id UUID REFERENCES news(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (news_id, tag_id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- Tabla: sync_status (Control de sincronización)
CREATE TABLE sync_status (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(100) NOT NULL,
    last_sync TIMESTAMP,
    records_synced INTEGER DEFAULT 0,
    status VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- Índices para optimizar búsquedas
-- ============================================
CREATE INDEX idx_news_url_hash ON news(url_hash);
CREATE INDEX idx_news_content_hash ON news(content_hash);
CREATE INDEX idx_news_published_date ON news(published_date);
CREATE INDEX idx_news_media_source ON news(media_source_id);
CREATE INDEX idx_news_multimedia_news ON news_multimedia(news_id);
CREATE INDEX idx_news_tags_news ON news_tags(news_id);
CREATE INDEX idx_news_tags_tag ON news_tags(tag_id);

-- ============================================
-- Datos iniciales
-- ============================================

-- Insertar medios de prensa del CSV
INSERT INTO media_sources (name, country) VALUES
    ('elsur', 'chile'),
    ('biobiochile', 'chile');

-- Tags predefinidos
INSERT INTO tags (name) VALUES
    ('política'),
    ('economía'),
    ('deportes'),
    ('cultura'),
    ('tecnología'),
    ('salud'),
    ('educación'),
    ('seguridad'),
    ('medio ambiente');

-- Estado de sincronización
INSERT INTO sync_status (service_name, status)
VALUES ('elasticsearch_sync', 'initialized');
