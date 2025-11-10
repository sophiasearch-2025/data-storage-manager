-- ===========================================================
-- 📦 ESTRUCTURA COMPLETA DE BASE DE DATOS: SISTEMA DE NOTICIAS
-- ===========================================================

-- (Opcional) Crear la base de datos si aún no existe:
-- CREATE DATABASE noticiasdb;
-- \c noticiasdb;

-- ===========================================================
-- 📰 TABLA: medios_prensa
-- ===========================================================
CREATE TABLE IF NOT EXISTS medios_prensa (
    id_medio SERIAL PRIMARY KEY,
    nombre VARCHAR(100) NOT NULL UNIQUE,
    url_main_page VARCHAR(255),
    pais VARCHAR(50),
    descripcion TEXT
);

-- ===========================================================
-- 🗞️ TABLA: noticia
-- ===========================================================
CREATE TABLE IF NOT EXISTS noticia (
    id_noticia SERIAL PRIMARY KEY,
    titulo VARCHAR(500) NOT NULL,
    url TEXT,
    fecha_subida DATE DEFAULT CURRENT_DATE,
    largo_noticia INT,
    id_medio INT NOT NULL,
    CONSTRAINT fk_medio
        FOREIGN KEY (id_medio)
        REFERENCES medios_prensa (id_medio)
        ON DELETE CASCADE
);

-- ===========================================================
-- 📄 TABLA: noticia_detalle
-- ===========================================================
CREATE TABLE IF NOT EXISTS noticia_detalle (
    id_detalle SERIAL PRIMARY KEY,
    texto_noticia TEXT NOT NULL,
    id_noticia INT UNIQUE,  -- Relación 1 a 1
    CONSTRAINT fk_noticia
        FOREIGN KEY (id_noticia)
        REFERENCES noticia (id_noticia)
        ON DELETE CASCADE
);

-- ===========================================================
-- 🏷️ TABLA: tag (etiquetas únicas)
-- ===========================================================
CREATE TABLE IF NOT EXISTS tag (
    id_tag SERIAL PRIMARY KEY,
    nombre VARCHAR(100) UNIQUE NOT NULL
);

-- ===========================================================
-- 🔗 TABLA INTERMEDIA: noticia_tag (relación muchos a muchos)
-- ===========================================================
CREATE TABLE IF NOT EXISTS noticia_tag (
    id_noticia INT NOT NULL,
    id_tag INT NOT NULL,
    PRIMARY KEY (id_noticia, id_tag),
    CONSTRAINT fk_noticia_tag_noticia
        FOREIGN KEY (id_noticia)
        REFERENCES noticia (id_noticia)
        ON DELETE CASCADE,
    CONSTRAINT fk_noticia_tag_tag
        FOREIGN KEY (id_tag)
        REFERENCES tag (id_tag)
        ON DELETE CASCADE
);

-- ===========================================================
-- ⚡ ÍNDICES RECOMENDADOS
-- ===========================================================
CREATE INDEX IF NOT EXISTS idx_medio_nombre ON medios_prensa (nombre);
CREATE INDEX IF NOT EXISTS idx_noticia_fecha ON noticia (fecha_subida);
CREATE INDEX IF NOT EXISTS idx_tag_nombre ON tag (nombre);

-- ===========================================================
-- 📋 VISTA OPCIONAL: noticias_completas
-- ===========================================================
CREATE OR REPLACE VIEW noticias_completas AS
SELECT 
    n.id_noticia,
    n.titulo,
    n.url,
    n.fecha_subida,
    m.nombre AS medio,
    m.pais,
    d.texto_noticia
FROM noticia n
JOIN medios_prensa m ON n.id_medio = m.id_medio
JOIN noticia_detalle d ON n.id_noticia = d.id_noticia;

-- ===========================================================
-- ✅ Comprobación final
-- ===========================================================
-- SELECT * FROM medios_prensa;
-- SELECT * FROM noticia;
-- SELECT * FROM noticia_detalle;
-- SELECT * FROM tag;
-- SELECT * FROM noticia_tag;
-- SELECT * FROM noticias_completas;
