CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE feeds (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name TEXT UNIQUE NOT NULL,
    url TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS articles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    link TEXT NOT NULL UNIQUE,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS share (
    id SERIAL PRIMARY KEY CHECK (id = 1),
    interval TEXT NOT NULL,
    workers_num INT NOT NULL
);

