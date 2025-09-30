-- Crear tabla de jugadores
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL,
    last_login TIMESTAMP NOT NULL
);

-- Crear tabla de mundos
CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    max_players INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL
);

-- Crear tabla de aldeas
CREATE TABLE IF NOT EXISTS villages (
    id UUID PRIMARY KEY,
    player_id UUID NOT NULL REFERENCES players(id),
    world_id UUID NOT NULL REFERENCES worlds(id),
    name VARCHAR(100) NOT NULL,
    x_coordinate INTEGER NOT NULL,
    y_coordinate INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL,
    UNIQUE(world_id, x_coordinate, y_coordinate)
);

-- Crear tabla de recursos
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY,
    village_id UUID NOT NULL REFERENCES villages(id),
    wood INTEGER NOT NULL DEFAULT 1000,
    stone INTEGER NOT NULL DEFAULT 1000,
    food INTEGER NOT NULL DEFAULT 1000,
    gold INTEGER NOT NULL DEFAULT 1000,
    last_updated TIMESTAMP NOT NULL,
    UNIQUE(village_id)
);

-- Crear tabla de edificios
CREATE TABLE IF NOT EXISTS buildings (
    id UUID PRIMARY KEY,
    village_id UUID NOT NULL REFERENCES villages(id),
    type VARCHAR(50) NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    is_upgrading BOOLEAN NOT NULL DEFAULT false,
    upgrade_completion_time TIMESTAMP,
    UNIQUE(village_id, type)
); 