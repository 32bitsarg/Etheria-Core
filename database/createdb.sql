-- ========================================
-- SCRIPT DE CREACIÓN DE BASE DE DATOS MMO
-- Basado en la estructura real del servidor
-- ========================================

-- Configuración inicial
SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

-- Crear extensión para UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ========================================
-- TABLAS PRINCIPALES
-- ========================================

-- Tabla de razas
CREATE TABLE IF NOT EXISTS races (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    trait_attack INTEGER DEFAULT 10 NOT NULL,
    trait_defense INTEGER DEFAULT 10 NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de mundos
CREATE TABLE IF NOT EXISTS worlds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    max_players INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    description TEXT DEFAULT '',
    current_players INTEGER DEFAULT 0,
    world_type VARCHAR(50) DEFAULT 'normal',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    is_online BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'offline',
    last_started_at TIMESTAMP WITH TIME ZONE,
    last_stopped_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de jugadores
CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    last_login TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    password_hash VARCHAR(255),
    race_id UUID REFERENCES races(id),
    level INTEGER DEFAULT 1,
    experience INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    is_online BOOLEAN DEFAULT false,
    last_active TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    role VARCHAR(20) DEFAULT 'user',
    is_banned BOOLEAN DEFAULT false,
    ban_reason TEXT,
    ban_expires_at TIMESTAMP WITH TIME ZONE,
    gold INTEGER DEFAULT 0,
    gems INTEGER DEFAULT 0,
    alliance_id UUID,
    world_id UUID REFERENCES worlds(id),
    CONSTRAINT players_role_check CHECK (role IN ('admin', 'moderator', 'user'))
);

-- Tabla de alianzas
CREATE TABLE IF NOT EXISTS alliances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    tag VARCHAR(10) NOT NULL,
    description TEXT,
    leader_id UUID NOT NULL REFERENCES players(id),
    world_id UUID NOT NULL REFERENCES worlds(id),
    member_count INTEGER DEFAULT 0,
    max_members INTEGER DEFAULT 50,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    level INTEGER DEFAULT 1,
    experience INTEGER DEFAULT 0
);

-- Tabla de aldeas
CREATE TABLE IF NOT EXISTS villages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id),
    name VARCHAR(100) NOT NULL,
    x_coordinate INTEGER NOT NULL,
    y_coordinate INTEGER NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- Tabla de recursos
CREATE TABLE IF NOT EXISTS resources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    village_id UUID NOT NULL REFERENCES villages(id) ON DELETE CASCADE,
    wood INTEGER DEFAULT 1000 NOT NULL,
    stone INTEGER DEFAULT 1000 NOT NULL,
    food INTEGER DEFAULT 1000 NOT NULL,
    gold INTEGER DEFAULT 1000 NOT NULL,
    last_updated TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

-- Tabla de tipos de edificios
CREATE TABLE IF NOT EXISTS building_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    max_level INTEGER DEFAULT 10,
    base_cost_wood INTEGER DEFAULT 0,
    base_cost_stone INTEGER DEFAULT 0,
    base_cost_food INTEGER DEFAULT 0,
    base_cost_gold INTEGER DEFAULT 0,
    cost_multiplier DECIMAL(5,2) DEFAULT 1.5,
    construction_time_base INTEGER DEFAULT 60,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de edificios
CREATE TABLE IF NOT EXISTS buildings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    village_id UUID NOT NULL REFERENCES villages(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    level INTEGER DEFAULT 1 NOT NULL,
    is_upgrading BOOLEAN DEFAULT false NOT NULL,
    upgrade_completion_time TIMESTAMP WITHOUT TIME ZONE,
    building_type_id UUID REFERENCES building_types(id)
);

-- Tabla de tipos de unidades
CREATE TABLE IF NOT EXISTS unit_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    attack INTEGER DEFAULT 10 NOT NULL,
    defense INTEGER DEFAULT 10 NOT NULL,
    speed INTEGER DEFAULT 10 NOT NULL,
    training_time INTEGER DEFAULT 60 NOT NULL,
    cost_wood INTEGER DEFAULT 0,
    cost_stone INTEGER DEFAULT 0,
    cost_food INTEGER DEFAULT 0,
    cost_gold INTEGER DEFAULT 0,
    food_consumption INTEGER DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de unidades del jugador
CREATE TABLE IF NOT EXISTS player_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    unit_type_id UUID NOT NULL REFERENCES unit_types(id),
    quantity INTEGER DEFAULT 0 NOT NULL,
    available_quantity INTEGER DEFAULT 0 NOT NULL,
    in_battle INTEGER DEFAULT 0 NOT NULL,
    training INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de tecnologías
CREATE TABLE IF NOT EXISTS technologies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    max_level INTEGER DEFAULT 10,
    base_cost_wood INTEGER DEFAULT 0,
    base_cost_stone INTEGER DEFAULT 0,
    base_cost_food INTEGER DEFAULT 0,
    base_cost_gold INTEGER DEFAULT 0,
    research_time_base INTEGER DEFAULT 300,
    effects JSONB,
    requirements JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de tecnologías del jugador
CREATE TABLE IF NOT EXISTS player_technologies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    technology_id UUID NOT NULL REFERENCES technologies(id),
    level INTEGER DEFAULT 0,
    is_researching BOOLEAN DEFAULT false,
    research_completion_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de héroes
CREATE TABLE IF NOT EXISTS heroes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    hero_type_id UUID,
    rarity VARCHAR(20) DEFAULT 'common',
    base_attack INTEGER DEFAULT 10,
    base_defense INTEGER DEFAULT 10,
    base_speed INTEGER DEFAULT 10,
    base_intelligence INTEGER DEFAULT 10,
    max_level INTEGER DEFAULT 50,
    experience_required JSONB,
    abilities JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de héroes del jugador
CREATE TABLE IF NOT EXISTS player_heroes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    hero_id UUID NOT NULL REFERENCES heroes(id),
    level INTEGER DEFAULT 1,
    experience INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de logros
CREATE TABLE IF NOT EXISTS achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    requirement_type VARCHAR(50) NOT NULL,
    requirement_value INTEGER NOT NULL,
    reward_type VARCHAR(50),
    reward_value INTEGER,
    icon_url VARCHAR(255),
    is_hidden BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de logros del jugador
CREATE TABLE IF NOT EXISTS player_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES achievements(id),
    progress INTEGER DEFAULT 0,
    is_completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de misiones
CREATE TABLE IF NOT EXISTS quests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL,
    requirement_type VARCHAR(50) NOT NULL,
    requirement_value INTEGER NOT NULL,
    reward_type VARCHAR(50),
    reward_value INTEGER,
    is_repeatable BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de misiones del jugador
CREATE TABLE IF NOT EXISTS player_quests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    quest_id UUID NOT NULL REFERENCES quests(id),
    progress INTEGER DEFAULT 0,
    is_completed BOOLEAN DEFAULT false,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de eventos
CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    event_type VARCHAR(50) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    rewards JSONB,
    requirements JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de categorías de eventos
CREATE TABLE IF NOT EXISTS event_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    color VARCHAR(7),
    background_color VARCHAR(7),
    display_order INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT true,
    show_in_calendar BOOLEAN DEFAULT true,
    show_in_dashboard BOOLEAN DEFAULT true,
    total_events INTEGER DEFAULT 0,
    active_events INTEGER DEFAULT 0,
    active_count INTEGER DEFAULT 0,
    total_participants INTEGER DEFAULT 0,
    completion_rate DECIMAL(5,2) DEFAULT 0.0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de eventos completos (reemplaza la tabla events básica)
CREATE TABLE IF NOT EXISTS game_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES event_categories(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    long_description TEXT,
    story_text TEXT,
    icon VARCHAR(100),
    color VARCHAR(7),
    background_color VARCHAR(7),
    banner_image VARCHAR(255),
    rarity VARCHAR(20) DEFAULT 'common',
    event_type VARCHAR(50) NOT NULL,
    event_format VARCHAR(50),
    max_participants INTEGER DEFAULT 0,
    min_participants INTEGER DEFAULT 1,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    registration_start TIMESTAMP WITH TIME ZONE,
    registration_end TIMESTAMP WITH TIME ZONE,
    duration INTEGER DEFAULT 0,
    level_required INTEGER DEFAULT 1,
    alliance_required UUID,
    prerequisites TEXT,
    entry_fee INTEGER DEFAULT 0,
    entry_currency VARCHAR(20) DEFAULT 'silver',
    event_rules TEXT,
    scoring_system TEXT,
    rewards_config TEXT,
    special_effects TEXT,
    status VARCHAR(20) DEFAULT 'upcoming',
    phase VARCHAR(20) DEFAULT 'registration',
    current_round INTEGER DEFAULT 1,
    total_rounds INTEGER DEFAULT 1,
    total_participants INTEGER DEFAULT 0,
    active_participants INTEGER DEFAULT 0,
    completion_rate DECIMAL(5,2) DEFAULT 0.0,
    average_score DECIMAL(10,2) DEFAULT 0.0,
    is_repeatable BOOLEAN DEFAULT false,
    repeat_interval VARCHAR(20),
    next_event_id UUID,
    is_hidden BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de participantes de eventos
CREATE TABLE IF NOT EXISTS event_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES game_events(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'registered',
    registration_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    entry_fee_paid BOOLEAN DEFAULT false,
    current_score INTEGER DEFAULT 0,
    total_score INTEGER DEFAULT 0,
    rank INTEGER DEFAULT 0,
    final_rank INTEGER DEFAULT 0,
    matches_played INTEGER DEFAULT 0,
    matches_won INTEGER DEFAULT 0,
    matches_lost INTEGER DEFAULT 0,
    matches_drawn INTEGER DEFAULT 0,
    rewards_earned BOOLEAN DEFAULT false,
    rewards_data TEXT,
    points_earned INTEGER DEFAULT 0,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    time_spent INTEGER DEFAULT 0,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    eliminated_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(event_id, player_id)
);

-- Tabla de partidas de eventos
CREATE TABLE IF NOT EXISTS event_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES game_events(id) ON DELETE CASCADE,
    round INTEGER NOT NULL,
    match_number INTEGER NOT NULL,
    player1_id UUID NOT NULL REFERENCES players(id),
    player2_id UUID NOT NULL REFERENCES players(id),
    winner_id UUID REFERENCES players(id),
    player1_score INTEGER DEFAULT 0,
    player2_score INTEGER DEFAULT 0,
    match_data TEXT,
    status VARCHAR(20) DEFAULT 'scheduled',
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    duration INTEGER DEFAULT 0,
    match_type VARCHAR(50),
    match_rules TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de recompensas de eventos
CREATE TABLE IF NOT EXISTS event_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES game_events(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(20) NOT NULL,
    min_rank INTEGER DEFAULT 1,
    max_rank INTEGER DEFAULT 999,
    min_score INTEGER DEFAULT 0,
    quantity INTEGER DEFAULT 1,
    resource_type VARCHAR(50),
    item_id UUID,
    currency_type VARCHAR(20),
    title_id UUID,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de estadísticas de eventos por jugador
CREATE TABLE IF NOT EXISTS event_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    total_events_joined INTEGER DEFAULT 0,
    active_events_joined INTEGER DEFAULT 0,
    completed_events INTEGER DEFAULT 0,
    events_won INTEGER DEFAULT 0,
    total_matches_played INTEGER DEFAULT 0,
    total_matches_won INTEGER DEFAULT 0,
    total_matches_lost INTEGER DEFAULT 0,
    win_rate DECIMAL(5,2) DEFAULT 0.0,
    total_score INTEGER DEFAULT 0,
    average_score DECIMAL(10,2) DEFAULT 0.0,
    highest_score INTEGER DEFAULT 0,
    total_rewards_earned INTEGER DEFAULT 0,
    total_points_earned INTEGER DEFAULT 0,
    total_time_spent INTEGER DEFAULT 0,
    first_place_finishes INTEGER DEFAULT 0,
    top_three_finishes INTEGER DEFAULT 0,
    perfect_scores INTEGER DEFAULT 0,
    first_event_date TIMESTAMP WITH TIME ZONE,
    last_event_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de notificaciones de eventos
CREATE TABLE IF NOT EXISTS event_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    event_id UUID NOT NULL REFERENCES game_events(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    data TEXT,
    is_read BOOLEAN DEFAULT false,
    is_dismissed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,
    dismissed_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de monedas
CREATE TABLE IF NOT EXISTS currencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    exchange_rate DECIMAL(10,4) DEFAULT 1.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de moneda global del jugador
CREATE TABLE IF NOT EXISTS player_global_currency (
    player_id UUID PRIMARY KEY REFERENCES players(id) ON DELETE CASCADE,
    amount BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de moneda del mundo del jugador
CREATE TABLE IF NOT EXISTS player_world_currency (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    world_id UUID NOT NULL REFERENCES worlds(id),
    currency_type VARCHAR(50) NOT NULL,
    amount BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de transacciones de moneda
CREATE TABLE IF NOT EXISTS currency_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    currency_type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    type VARCHAR(20) NOT NULL,
    description TEXT,
    balance BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de canales de chat
CREATE TABLE IF NOT EXISTS chat_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'global' NOT NULL,
    world_id UUID REFERENCES worlds(id),
    is_active BOOLEAN DEFAULT true NOT NULL,
    max_members INTEGER DEFAULT 1000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de mensajes de chat
CREATE TABLE IF NOT EXISTS chat_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES chat_channels(id) ON DELETE CASCADE,
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    username VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'text' NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de notificaciones
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_read BOOLEAN DEFAULT false,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de administradores
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    permissions JSONB,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de batallas
CREATE TABLE IF NOT EXISTS battles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attacker_id UUID NOT NULL REFERENCES players(id),
    defender_id UUID NOT NULL REFERENCES players(id),
    attacker_village_id UUID NOT NULL REFERENCES villages(id),
    defender_village_id UUID NOT NULL REFERENCES villages(id),
    status VARCHAR(20) DEFAULT 'pending',
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    result JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de rankings
CREATE TABLE IF NOT EXISTS ranking_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    sub_type VARCHAR(50),
    icon VARCHAR(100),
    color VARCHAR(7),
    update_interval INTEGER DEFAULT 60,
    max_positions INTEGER DEFAULT 1000,
    min_score INTEGER DEFAULT 0,
    score_formula TEXT,
    rewards_enabled BOOLEAN DEFAULT false,
    rewards_config TEXT,
    display_order INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT true,
    show_in_dashboard BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    last_updated TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabla de entradas de ranking
CREATE TABLE IF NOT EXISTS ranking_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id INTEGER NOT NULL REFERENCES ranking_categories(id),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    score BIGINT NOT NULL,
    rank INTEGER,
    data JSONB,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ========================================
-- ÍNDICES PARA OPTIMIZACIÓN
-- ========================================

-- Índices para jugadores
CREATE INDEX IF NOT EXISTS idx_players_username ON players(username);
CREATE INDEX IF NOT EXISTS idx_players_email ON players(email);
CREATE INDEX IF NOT EXISTS idx_players_world_id ON players(world_id);
CREATE INDEX IF NOT EXISTS idx_players_alliance_id ON players(alliance_id);

-- Índices para aldeas
CREATE INDEX IF NOT EXISTS idx_villages_player_id ON villages(player_id);
CREATE INDEX IF NOT EXISTS idx_villages_world_id ON villages(world_id);
CREATE INDEX IF NOT EXISTS idx_villages_coordinates ON villages(x_coordinate, y_coordinate);

-- Índices para recursos
CREATE INDEX IF NOT EXISTS idx_resources_village_id ON resources(village_id);

-- Índices para edificios
CREATE INDEX IF NOT EXISTS idx_buildings_village_id ON buildings(village_id);
CREATE INDEX IF NOT EXISTS idx_buildings_type ON buildings(type);

-- Índices para unidades
CREATE INDEX IF NOT EXISTS idx_player_units_player_id ON player_units(player_id);
CREATE INDEX IF NOT EXISTS idx_player_units_unit_type_id ON player_units(unit_type_id);

-- Índices para tecnologías
CREATE INDEX IF NOT EXISTS idx_player_technologies_player_id ON player_technologies(player_id);
CREATE INDEX IF NOT EXISTS idx_player_technologies_technology_id ON player_technologies(technology_id);

-- Índices para héroes
CREATE INDEX IF NOT EXISTS idx_player_heroes_player_id ON player_heroes(player_id);
CREATE INDEX IF NOT EXISTS idx_player_heroes_hero_id ON player_heroes(hero_id);

-- Índices para logros
CREATE INDEX IF NOT EXISTS idx_player_achievements_player_id ON player_achievements(player_id);
CREATE INDEX IF NOT EXISTS idx_player_achievements_achievement_id ON player_achievements(achievement_id);

-- Índices para misiones
CREATE INDEX IF NOT EXISTS idx_player_quests_player_id ON player_quests(player_id);
CREATE INDEX IF NOT EXISTS idx_player_quests_quest_id ON player_quests(quest_id);

-- Índices para chat
CREATE INDEX IF NOT EXISTS idx_chat_messages_channel_id ON chat_messages(channel_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_player_id ON chat_messages(player_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages(created_at);

-- Índices para notificaciones
CREATE INDEX IF NOT EXISTS idx_notifications_player_id ON notifications(player_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);

-- Índices para batallas
CREATE INDEX IF NOT EXISTS idx_battles_attacker_id ON battles(attacker_id);
CREATE INDEX IF NOT EXISTS idx_battles_defender_id ON battles(defender_id);
CREATE INDEX IF NOT EXISTS idx_battles_status ON battles(status);

-- Índices para rankings
CREATE INDEX IF NOT EXISTS idx_ranking_entries_category_id ON ranking_entries(category_id);
CREATE INDEX IF NOT EXISTS idx_ranking_entries_player_id ON ranking_entries(player_id);
CREATE INDEX IF NOT EXISTS idx_ranking_entries_score ON ranking_entries(score);

-- ========================================
-- DATOS INICIALES
-- ========================================

-- Insertar razas básicas
INSERT INTO races (name, trait_attack, trait_defense, description) VALUES
('Humanos', 12, 8, 'Raza equilibrada con bonificación de ataque'),
('Elfos', 10, 10, 'Raza mágica con bonificación de defensa'),
('Orcos', 15, 5, 'Raza guerrera con gran poder de ataque'),
('Enanos', 8, 15, 'Raza resistente con gran defensa')
ON CONFLICT DO NOTHING;

-- Insertar tipos de edificios básicos
INSERT INTO building_types (name, display_name, description, max_level, base_cost_wood, base_cost_stone, base_cost_food, base_cost_gold, cost_multiplier, construction_time_base) VALUES
('town_hall', 'Ayuntamiento', 'Centro de la aldea', 20, 100, 100, 50, 0, 1.5, 120),
('warehouse', 'Almacén', 'Almacena recursos', 20, 80, 80, 40, 0, 1.5, 90),
('granary', 'Granero', 'Almacena comida', 20, 80, 80, 40, 0, 1.5, 90),
('marketplace', 'Mercado', 'Comercio con otros jugadores', 20, 120, 100, 60, 0, 1.5, 150),
('barracks', 'Cuartel', 'Entrena unidades militares', 20, 150, 120, 80, 0, 1.5, 180),
('wood_cutter', 'Leñador', 'Produce madera', 20, 60, 40, 30, 0, 1.5, 60),
('stone_quarry', 'Cantera', 'Produce piedra', 20, 60, 40, 30, 0, 1.5, 60),
('farm', 'Granja', 'Produce comida', 20, 60, 40, 30, 0, 1.5, 60),
('gold_mine', 'Mina de Oro', 'Produce oro', 20, 100, 80, 50, 0, 1.5, 120)
ON CONFLICT DO NOTHING;

-- Insertar tipos de unidades básicas
INSERT INTO unit_types (name, display_name, description, attack, defense, speed, training_time, cost_wood, cost_stone, cost_food, cost_gold, food_consumption) VALUES
('swordsman', 'Espadachín', 'Unidad básica de infantería', 15, 10, 8, 120, 50, 30, 20, 0, 1),
('archer', 'Arquero', 'Unidad de ataque a distancia', 20, 5, 10, 180, 60, 40, 25, 0, 1),
('cavalry', 'Caballería', 'Unidad rápida de ataque', 25, 8, 15, 300, 80, 60, 40, 0, 2),
('defender', 'Defensor', 'Unidad especializada en defensa', 8, 20, 6, 150, 70, 50, 30, 0, 1)
ON CONFLICT DO NOTHING;

-- Insertar tecnologías básicas
INSERT INTO technologies (name, description, category, max_level, base_cost_wood, base_cost_stone, base_cost_food, base_cost_gold, research_time_base, effects, requirements) VALUES
('wood_production', 'Aumenta la producción de madera', 'economy', 20, 100, 50, 30, 0, 300, '{"wood_production_bonus": 10}', '{"town_hall_level": 1}'),
('stone_production', 'Aumenta la producción de piedra', 'economy', 20, 100, 50, 30, 0, 300, '{"stone_production_bonus": 10}', '{"town_hall_level": 1}'),
('food_production', 'Aumenta la producción de comida', 'economy', 20, 100, 50, 30, 0, 300, '{"food_production_bonus": 10}', '{"town_hall_level": 1}'),
('military_tactics', 'Mejora el ataque de las unidades', 'military', 20, 150, 100, 80, 0, 600, '{"unit_attack_bonus": 5}', '{"barracks_level": 3}'),
('defense_systems', 'Mejora la defensa de las unidades', 'military', 20, 150, 100, 80, 0, 600, '{"unit_defense_bonus": 5}', '{"barracks_level": 3}')
ON CONFLICT DO NOTHING;

-- Insertar logros básicos
INSERT INTO achievements (name, description, category, requirement_type, requirement_value, reward_type, reward_value, icon_url) VALUES
('Primera Aldea', 'Construye tu primera aldea', 'building', 'village_count', 1, 'gold', 100, '/icons/first_village.png'),
('Constructor', 'Mejora un edificio al nivel 5', 'building', 'building_level', 5, 'experience', 50, '/icons/constructor.png'),
('Recolector', 'Acumula 10,000 recursos', 'economy', 'total_resources', 10000, 'gold', 200, '/icons/collector.png'),
('Guerrero', 'Entrena 100 unidades', 'military', 'units_trained', 100, 'experience', 100, '/icons/warrior.png')
ON CONFLICT DO NOTHING;

-- Insertar misiones básicas
INSERT INTO quests (name, description, category, requirement_type, requirement_value, reward_type, reward_value, is_repeatable) VALUES
('Bienvenido', 'Completa el tutorial básico', 'tutorial', 'tutorial_completed', 1, 'gold', 500, false),
('Constructor Novato', 'Mejora 3 edificios diferentes', 'building', 'different_buildings_upgraded', 3, 'experience', 100, false),
('Recolector de Recursos', 'Recolecta 5,000 recursos', 'economy', 'resources_collected', 5000, 'gold', 150, true),
('Entrenador de Unidades', 'Entrena 50 unidades', 'military', 'units_trained', 50, 'experience', 75, true)
ON CONFLICT DO NOTHING;

-- Insertar canales de chat básicos
INSERT INTO chat_channels (name, description, type, is_active, max_members) VALUES
('Global', 'Chat global para todos los jugadores', 'global', true, 10000),
('Ayuda', 'Canal de ayuda para nuevos jugadores', 'help', true, 1000),
('Comercio', 'Canal para intercambios y comercio', 'trade', true, 5000)
ON CONFLICT DO NOTHING;

-- Insertar monedas básicas
INSERT INTO currencies (name, symbol, description, is_active, exchange_rate) VALUES
('Oro', 'G', 'Moneda básica del juego', true, 1.0),
('Gemas', '💎', 'Moneda premium del juego', true, 100.0),
('Moneda Global', '🌍', 'Moneda para compras globales', true, 1.0)
ON CONFLICT DO NOTHING;

-- Insertar categorías de ranking básicas
INSERT INTO ranking_categories (name, description, type, sub_type, icon, color, update_interval, max_positions, score_formula, rewards_enabled, display_order, is_public, show_in_dashboard) VALUES
('Nivel de Jugador', 'Ranking por nivel de jugador', 'player', 'level', '/icons/level.png', '#FFD700', 300, 1000, '{"formula": "level * 100 + experience"}', true, 1, true, true),
('Poder Militar', 'Ranking por poder militar total', 'military', 'power', '/icons/military.png', '#FF4500', 600, 1000, '{"formula": "total_attack + total_defense"}', true, 2, true, true),
('Economía', 'Ranking por riqueza total', 'economy', 'wealth', '/icons/economy.png', '#32CD32', 900, 1000, '{"formula": "total_resources + gold * 10"}', true, 3, true, true),
('Construcción', 'Ranking por nivel de edificios', 'building', 'construction', '/icons/building.png', '#4169E1', 1200, 1000, '{"formula": "sum(building_levels) * 10"}', true, 4, true, true)
ON CONFLICT DO NOTHING;

-- ========================================
-- FUNCIONES ÚTILES
-- ========================================

-- Función para actualizar timestamp de actualización
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Función para agregar moneda global
CREATE OR REPLACE FUNCTION add_global_currency(p_player_id UUID, p_amount BIGINT, p_description VARCHAR DEFAULT NULL)
RETURNS BIGINT AS $$
DECLARE
    new_balance BIGINT;
BEGIN
    INSERT INTO player_global_currency (player_id, amount)
    VALUES (p_player_id, p_amount)
    ON CONFLICT (player_id)
    DO UPDATE SET amount = player_global_currency.amount + p_amount, updated_at = NOW()
    RETURNING amount INTO new_balance;
    
    -- Registrar transacción
    INSERT INTO currency_transactions (player_id, currency_type, amount, type, description, balance)
    VALUES (p_player_id, 'global', p_amount, 'earn', p_description, new_balance);
    
    RETURN new_balance;
END;
$$ LANGUAGE plpgsql;

-- Función para agregar puntos de prestigio
CREATE OR REPLACE FUNCTION add_prestige_points(p_player_id UUID, p_points INTEGER, p_source VARCHAR DEFAULT 'other')
RETURNS BOOLEAN AS $$
DECLARE
    current_prestige RECORD;
    new_level INTEGER;
    level_threshold INTEGER;
BEGIN
    -- Obtener prestigio actual del jugador
    SELECT * INTO current_prestige
    FROM player_prestige
    WHERE player_id = p_player_id;
    
    -- Si no existe, crear registro
    IF NOT FOUND THEN
        INSERT INTO player_prestige (player_id, prestige_level, prestige_points, total_prestige_earned)
        VALUES (p_player_id, 0, p_points, p_points);
        RETURN TRUE;
    END IF;
    
    -- Calcular nuevo nivel (cada 100 puntos = 1 nivel)
    new_level := (current_prestige.prestige_points + p_points) / 100;
    level_threshold := new_level * 100;
    
    -- Actualizar prestigio
    UPDATE player_prestige
    SET 
        prestige_points = prestige_points + p_points,
        prestige_level = new_level,
        total_prestige_earned = total_prestige_earned + p_points,
        updated_at = NOW()
    WHERE player_id = p_player_id;
    
    RETURN TRUE;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- FUNCIONES AVANZADAS Y MEJORADAS
-- ========================================

-- Función avanzada para calcular producción de recursos (REEMPLAZA LA ANTERIOR)
CREATE OR REPLACE FUNCTION calculate_resource_production_advanced(p_village_id UUID)
RETURNS TABLE(wood_production INTEGER, stone_production INTEGER, food_production INTEGER, gold_production INTEGER) AS $$
DECLARE
    base_wood INTEGER;
    base_stone INTEGER;
    base_food INTEGER;
    base_gold INTEGER;
    tech_bonus DECIMAL(5,2);
    alliance_bonus DECIMAL(5,2);
    event_bonus DECIMAL(5,2);
    world_bonus DECIMAL(5,2);
    player_id UUID;
    alliance_id UUID;
    world_id UUID;
BEGIN
    -- Obtener información básica de la aldea
    SELECT v.player_id, v.world_id, p.alliance_id INTO player_id, world_id, alliance_id
    FROM villages v
    JOIN players p ON v.player_id = p.id
    WHERE v.id = p_village_id;
    
    -- Calcular producción base por edificios
    SELECT 
        COALESCE(SUM(CASE WHEN b.type = 'wood_cutter' THEN b.level * 10 ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN b.type = 'stone_quarry' THEN b.level * 10 ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN b.type = 'farm' THEN b.level * 10 ELSE 0 END), 0),
        COALESCE(SUM(CASE WHEN b.type = 'gold_mine' THEN b.level * 5 ELSE 0 END), 0)
    INTO base_wood, base_stone, base_food, base_gold
    FROM buildings b
    WHERE b.village_id = p_village_id;
    
    -- Calcular bonificación por tecnologías
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN pt.technology_id = (SELECT id FROM technologies WHERE name = 'wood_production') THEN pt.level * 0.1
                WHEN pt.technology_id = (SELECT id FROM technologies WHERE name = 'stone_production') THEN pt.level * 0.1
                WHEN pt.technology_id = (SELECT id FROM technologies WHERE name = 'food_production') THEN pt.level * 0.1
                ELSE 0
            END
        ), 0) INTO tech_bonus
    FROM player_technologies pt
    WHERE pt.player_id = player_id;
    
    -- Calcular bonificación por alianza
    SELECT 
        COALESCE(a.level * 0.05, 0) INTO alliance_bonus
    FROM alliances a
    WHERE a.id = alliance_id AND a.is_active = true;
    
    -- Calcular bonificación por eventos activos
    SELECT 
        COALESCE(SUM(
            CASE 
                WHEN e.event_type = 'resource_boost' THEN 0.2
                WHEN e.event_type = 'production_festival' THEN 0.15
                ELSE 0
            END
        ), 0) INTO event_bonus
    FROM events e
    WHERE e.is_active = true 
    AND NOW() BETWEEN e.start_time AND e.end_time;
    
    -- Calcular bonificación por tipo de mundo
    SELECT 
        CASE 
            WHEN w.world_type = 'peaceful' THEN 0.1
            WHEN w.world_type = 'pvp' THEN 0.05
            ELSE 0
        END INTO world_bonus
    FROM worlds w
    WHERE w.id = world_id;
    
    -- Aplicar todas las bonificaciones
    RETURN QUERY
    SELECT 
        (base_wood * (1 + tech_bonus + alliance_bonus + event_bonus + world_bonus))::INTEGER,
        (base_stone * (1 + tech_bonus + alliance_bonus + event_bonus + world_bonus))::INTEGER,
        (base_food * (1 + tech_bonus + alliance_bonus + event_bonus + world_bonus))::INTEGER,
        (base_gold * (1 + tech_bonus + alliance_bonus + event_bonus + world_bonus))::INTEGER;
END;
$$ LANGUAGE plpgsql;

-- Función avanzada para verificar requisitos de construcción (REEMPLAZA LA ANTERIOR)
CREATE OR REPLACE FUNCTION check_building_requirements_advanced(p_village_id UUID, p_building_type VARCHAR, p_target_level INTEGER DEFAULT 1)
RETURNS TABLE(can_build BOOLEAN, missing_requirements TEXT[], cost_wood INTEGER, cost_stone INTEGER, cost_food INTEGER, cost_gold INTEGER) AS $$
DECLARE
    player_id UUID;
    world_id UUID;
    alliance_id UUID;
    town_hall_level INTEGER;
    required_town_hall INTEGER;
    required_techs TEXT[];
    required_buildings TEXT[];
    missing_techs TEXT[];
    missing_buildings TEXT[];
    building_cost RECORD;
    current_resources RECORD;
    total_cost_wood INTEGER;
    total_cost_stone INTEGER;
    total_cost_food INTEGER;
    total_cost_gold INTEGER;
BEGIN
    -- Obtener información básica
    SELECT v.player_id, v.world_id, p.alliance_id INTO player_id, world_id, alliance_id
    FROM villages v
    JOIN players p ON v.player_id = p.id
    WHERE v.id = p_village_id;
    
    -- Obtener nivel del ayuntamiento
    SELECT level INTO town_hall_level
    FROM buildings
    WHERE village_id = p_village_id AND type = 'town_hall';
    
    -- Definir requisitos según el tipo de edificio
    CASE p_building_type
        WHEN 'warehouse', 'granary' THEN
            required_town_hall := 1;
            required_techs := ARRAY['basic_construction'];
            required_buildings := ARRAY['town_hall'];
        WHEN 'marketplace' THEN
            required_town_hall := 3;
            required_techs := ARRAY['basic_construction', 'trade_basics'];
            required_buildings := ARRAY['town_hall', 'warehouse'];
        WHEN 'barracks' THEN
            required_town_hall := 5;
            required_techs := ARRAY['basic_construction', 'military_basics'];
            required_buildings := ARRAY['town_hall', 'warehouse'];
        WHEN 'wood_cutter', 'stone_quarry', 'farm', 'gold_mine' THEN
            required_town_hall := 2;
            required_techs := ARRAY['basic_construction', 'resource_management'];
            required_buildings := ARRAY['town_hall'];
        ELSE
            required_town_hall := 1;
            required_techs := ARRAY['basic_construction'];
            required_buildings := ARRAY['town_hall'];
    END CASE;
    
    -- Verificar requisitos de ayuntamiento
    IF COALESCE(town_hall_level, 0) < required_town_hall THEN
        missing_buildings := array_append(missing_buildings, 
            format('Ayuntamiento nivel %s (actual: %s)', required_town_hall, COALESCE(town_hall_level, 0)));
    END IF;
    
    -- Verificar tecnologías requeridas
    SELECT array_agg(t.name) INTO missing_techs
    FROM unnest(required_techs) AS t(name)
    WHERE NOT EXISTS (
        SELECT 1 FROM player_technologies pt
        JOIN technologies tech ON pt.technology_id = tech.id
        WHERE pt.player_id = player_id AND tech.name = t.name AND pt.level > 0
    );
    
    -- Verificar edificios requeridos
    SELECT array_agg(b.name) INTO missing_buildings
    FROM unnest(required_buildings) AS b(name)
    WHERE NOT EXISTS (
        SELECT 1 FROM buildings bld
        WHERE bld.village_id = p_village_id AND bld.type = b.name AND bld.level > 0
    );
    
    -- Calcular costo del edificio
    SELECT 
        base_cost_wood * power(cost_multiplier, p_target_level - 1),
        base_cost_stone * power(cost_multiplier, p_target_level - 1),
        base_cost_food * power(cost_multiplier, p_target_level - 1),
        base_cost_gold * power(cost_multiplier, p_target_level - 1)
    INTO building_cost
    FROM building_types
    WHERE name = p_building_type;
    
    -- Obtener recursos actuales
    SELECT wood, stone, food, gold INTO current_resources
    FROM resources
    WHERE village_id = p_village_id;
    
    -- Calcular costos totales
    total_cost_wood := COALESCE(building_cost.base_cost_wood, 0);
    total_cost_stone := COALESCE(building_cost.base_cost_stone, 0);
    total_cost_food := COALESCE(building_cost.base_cost_food, 0);
    total_cost_gold := COALESCE(building_cost.base_cost_gold, 0);
    
    -- Verificar si hay recursos suficientes
    IF COALESCE(current_resources.wood, 0) < total_cost_wood THEN
        missing_requirements := array_append(missing_requirements, 
            format('Madera: %s (disponible: %s)', total_cost_wood, COALESCE(current_resources.wood, 0)));
    END IF;
    
    IF COALESCE(current_resources.stone, 0) < total_cost_stone THEN
        missing_requirements := array_append(missing_requirements, 
            format('Piedra: %s (disponible: %s)', total_cost_stone, COALESCE(current_resources.stone, 0)));
    END IF;
    
    IF COALESCE(current_resources.food, 0) < total_cost_food THEN
        missing_requirements := array_append(missing_requirements, 
            format('Comida: %s (disponible: %s)', total_cost_food, COALESCE(current_resources.food, 0)));
    END IF;
    
    IF COALESCE(current_resources.gold, 0) < total_cost_gold THEN
        missing_requirements := array_append(missing_requirements, 
            format('Oro: %s (disponible: %s)', total_cost_gold, COALESCE(current_resources.gold, 0)));
    END IF;
    
    -- Combinar todos los requisitos faltantes
    missing_requirements := array_cat(missing_requirements, missing_techs);
    missing_requirements := array_cat(missing_requirements, missing_buildings);
    
    -- Determinar si se puede construir
    RETURN QUERY
    SELECT 
        array_length(missing_requirements, 1) IS NULL OR array_length(missing_requirements, 1) = 0,
        COALESCE(missing_requirements, ARRAY[]::TEXT[]),
        total_cost_wood,
        total_cost_stone,
        total_cost_food,
        total_cost_gold;
END;
$$ LANGUAGE plpgsql;

-- Función para calcular resultado de batalla
CREATE OR REPLACE FUNCTION calculate_battle_outcome(
    p_attacker_units JSONB,
    p_defender_units JSONB,
    p_attacker_heroes JSONB DEFAULT '[]',
    p_defender_heroes JSONB DEFAULT '[]',
    p_terrain VARCHAR DEFAULT 'plains',
    p_weather VARCHAR DEFAULT 'clear',
    p_attacker_technologies JSONB DEFAULT '[]',
    p_defender_technologies JSONB DEFAULT '[]'
)
RETURNS TABLE(
    attacker_victory BOOLEAN,
    attacker_losses JSONB,
    defender_losses JSONB,
    battle_duration INTEGER,
    experience_gained INTEGER,
    resources_plundered JSONB
) AS $$
DECLARE
    attacker_power INTEGER;
    defender_power INTEGER;
    terrain_modifier DECIMAL(3,2);
    weather_modifier DECIMAL(3,2);
    hero_bonus DECIMAL(3,2);
    tech_bonus DECIMAL(3,2);
    battle_rounds INTEGER;
    attacker_losses JSONB;
    defender_losses JSONB;
    experience_gained INTEGER;
    resources_plundered JSONB;
BEGIN
    -- Calcular poder base de las unidades
    SELECT 
        COALESCE(SUM((unit->>'quantity')::INTEGER * (unit->>'attack')::INTEGER), 0),
        COALESCE(SUM((unit->>'quantity')::INTEGER * (unit->>'defense')::INTEGER), 0)
    INTO attacker_power, defender_power
    FROM jsonb_array_elements(p_attacker_units) AS unit;
    
    -- Aplicar modificadores de terreno
    CASE p_terrain
        WHEN 'forest' THEN terrain_modifier := 1.1;  -- Bonificación para unidades de madera
        WHEN 'mountain' THEN terrain_modifier := 1.15; -- Bonificación para unidades de piedra
        WHEN 'water' THEN terrain_modifier := 0.9;   -- Penalización general
        WHEN 'desert' THEN terrain_modifier := 0.85; -- Penalización general
        ELSE terrain_modifier := 1.0; -- Llanuras
    END CASE;
    
    -- Aplicar modificadores de clima
    CASE p_weather
        WHEN 'rain' THEN weather_modifier := 0.9;   -- Penalización por lluvia
        WHEN 'storm' THEN weather_modifier := 0.8;  -- Penalización por tormenta
        WHEN 'fog' THEN weather_modifier := 0.95;   -- Penalización ligera por niebla
        ELSE weather_modifier := 1.0; -- Claro
    END CASE;
    
    -- Calcular bonificación por héroes
    SELECT COALESCE(SUM((hero->>'level')::INTEGER * 0.05), 0) INTO hero_bonus
    FROM jsonb_array_elements(p_attacker_heroes) AS hero;
    
    -- Calcular bonificación por tecnologías
    SELECT COALESCE(SUM((tech->>'level')::INTEGER * 0.02), 0) INTO tech_bonus
    FROM jsonb_array_elements(p_attacker_technologies) AS tech;
    
    -- Aplicar todos los modificadores
    attacker_power := (attacker_power * terrain_modifier * weather_modifier * (1 + hero_bonus + tech_bonus))::INTEGER;
    
    -- Calcular duración de la batalla (mínimo 3 rondas, máximo 10)
    battle_rounds := GREATEST(3, LEAST(10, (GREATEST(attacker_power, defender_power) / 100)));
    
    -- Determinar victoria basada en poder y aleatoriedad
    IF attacker_power > defender_power * 1.5 THEN
        -- Victoria aplastante
        attacker_victory := true;
        attacker_losses := jsonb_build_object('total', FLOOR(attacker_power * 0.1));
        defender_losses := jsonb_build_object('total', FLOOR(defender_power * 0.8));
    ELSIF attacker_power > defender_power THEN
        -- Victoria por estrecho margen
        attacker_victory := true;
        attacker_losses := jsonb_build_object('total', FLOOR(attacker_power * 0.3));
        defender_losses := jsonb_build_object('total', FLOOR(defender_power * 0.6));
    ELSE
        -- Derrota
        attacker_victory := false;
        attacker_losses := jsonb_build_object('total', FLOOR(attacker_power * 0.7));
        defender_losses := jsonb_build_object('total', FLOOR(defender_power * 0.2));
    END IF;
    
    -- Calcular experiencia ganada
    experience_gained := CASE 
        WHEN attacker_victory THEN 
            GREATEST(10, LEAST(100, FLOOR(defender_power / 10)))
        ELSE 
            GREATEST(5, LEAST(50, FLOOR(attacker_power / 20)))
    END;
    
    -- Calcular recursos saqueados (solo si gana el atacante)
    IF attacker_victory THEN
        resources_plundered := jsonb_build_object(
            'wood', FLOOR(RANDOM() * 100 + 50),
            'stone', FLOOR(RANDOM() * 80 + 40),
            'food', FLOOR(RANDOM() * 60 + 30),
            'gold', FLOOR(RANDOM() * 20 + 10)
        );
    ELSE
        resources_plundered := jsonb_build_object('wood', 0, 'stone', 0, 'food', 0, 'gold', 0);
    END IF;
    
    RETURN QUERY
    SELECT 
        attacker_victory,
        attacker_losses,
        defender_losses,
        battle_rounds,
        experience_gained,
        resources_plundered;
END;
$$ LANGUAGE plpgsql;

-- Función para procesar cola de construcción
CREATE OR REPLACE FUNCTION process_construction_queue(p_village_id UUID)
RETURNS TABLE(
    building_type VARCHAR,
    old_level INTEGER,
    new_level INTEGER,
    construction_time INTEGER,
    resources_spent JSONB
) AS $$
DECLARE
    building_record RECORD;
    building_config RECORD;
    construction_bonus DECIMAL(3,2);
    actual_construction_time INTEGER;
    resources_spent JSONB;
BEGIN
    -- Obtener edificios en construcción
    FOR building_record IN
        SELECT b.id, b.type, b.level, b.upgrade_completion_time
        FROM buildings b
        WHERE b.village_id = p_village_id 
        AND b.is_upgrading = true 
        AND b.upgrade_completion_time <= NOW()
    LOOP
        -- Obtener configuración del edificio
        SELECT 
            base_cost_wood * power(cost_multiplier, building_record.level),
            base_cost_stone * power(cost_multiplier, building_record.level),
            base_cost_food * power(cost_multiplier, building_record.level),
            base_cost_gold * power(cost_multiplier, building_record.level),
            construction_time_base * power(1.2, building_record.level - 1)
        INTO building_config
        FROM building_types
        WHERE name = building_record.type;
        
        -- Calcular bonificaciones de construcción
        SELECT 
            COALESCE(SUM(
                CASE 
                    WHEN pt.technology_id = (SELECT id FROM technologies WHERE name = 'construction_speed') 
                    THEN pt.level * 0.05
                    ELSE 0
                END
            ), 0) INTO construction_bonus
        FROM player_technologies pt
        JOIN players p ON pt.player_id = p.id
        JOIN villages v ON p.id = v.player_id
        WHERE v.id = p_village_id;
        
        -- Calcular tiempo real de construcción
        actual_construction_time := (building_config.construction_time_base * (1 - construction_bonus))::INTEGER;
        
        -- Calcular recursos gastados
        resources_spent := jsonb_build_object(
            'wood', building_config.base_cost_wood,
            'stone', building_config.base_cost_stone,
            'food', building_config.base_cost_food,
            'gold', building_config.base_cost_gold
        );
        
        -- Completar la construcción
        UPDATE buildings 
        SET 
            level = level + 1,
            is_upgrading = false,
            upgrade_completion_time = NULL,
            updated_at = NOW()
        WHERE id = building_record.id;
        
        -- Retornar información de la mejora completada
        building_type := building_record.type;
        old_level := building_record.level;
        new_level := building_record.level + 1;
        construction_time := actual_construction_time;
        
        RETURN NEXT;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Función para calcular tasas de intercambio
CREATE OR REPLACE FUNCTION calculate_trade_rates(
    p_resource_type VARCHAR,
    p_world_id UUID DEFAULT NULL
)
RETURNS TABLE(
    resource_type VARCHAR,
    current_price INTEGER,
    price_trend VARCHAR,
    supply_level VARCHAR,
    demand_level VARCHAR,
    recommended_action VARCHAR
) AS $$
DECLARE
    avg_price INTEGER;
    price_24h_ago INTEGER;
    price_7d_ago INTEGER;
    supply_quantity INTEGER;
    demand_quantity INTEGER;
    price_change_24h DECIMAL(5,2);
    price_change_7d DECIMAL(5,2);
    trend VARCHAR;
    supply_level VARCHAR;
    demand_level VARCHAR;
    action VARCHAR;
BEGIN
    -- Obtener precio actual promedio
    SELECT COALESCE(AVG(price_per_unit), 0) INTO avg_price
    FROM market_listings
    WHERE resource_type = p_resource_type 
    AND (p_world_id IS NULL OR world_id = p_world_id)
    AND status = 'active';
    
    -- Obtener precio hace 24 horas
    SELECT COALESCE(AVG(price_per_unit), avg_price) INTO price_24h_ago
    FROM market_listings
    WHERE resource_type = p_resource_type 
    AND (p_world_id IS NULL OR world_id = p_world_id)
    AND created_at >= NOW() - INTERVAL '24 hours';
    
    -- Obtener precio hace 7 días
    SELECT COALESCE(AVG(price_per_unit), avg_price) INTO price_7d_ago
    FROM market_listings
    WHERE resource_type = p_resource_type 
    AND (p_world_id IS NULL OR world_id = p_world_id)
    AND created_at >= NOW() - INTERVAL '7 days';
    
    -- Calcular cambios de precio
    price_change_24h := CASE 
        WHEN price_24h_ago > 0 THEN ((avg_price - price_24h_ago) / price_24h_ago * 100)
        ELSE 0
    END;
    
    price_change_7d := CASE 
        WHEN price_7d_ago > 0 THEN ((avg_price - price_7d_ago) / price_7d_ago * 100)
        ELSE 0
    END;
    
    -- Determinar tendencia
    IF price_change_24h > 5 THEN
        trend := 'rising';
    ELSIF price_change_24h < -5 THEN
        trend := 'falling';
    ELSE
        trend := 'stable';
    END IF;
    
    -- Calcular nivel de oferta
    SELECT COALESCE(SUM(quantity), 0) INTO supply_quantity
    FROM market_listings
    WHERE resource_type = p_resource_type 
    AND (p_world_id IS NULL OR world_id = p_world_id)
    AND status = 'active'
    AND offer_type = 'sell';
    
    -- Calcular nivel de demanda
    SELECT COALESCE(SUM(quantity), 0) INTO demand_quantity
    FROM market_listings
    WHERE resource_type = p_resource_type 
    AND (p_world_id IS NULL OR world_id = p_world_id)
    AND status = 'active'
    AND offer_type = 'buy';
    
    -- Determinar niveles de oferta y demanda
    supply_level := CASE 
        WHEN supply_quantity > demand_quantity * 2 THEN 'high'
        WHEN supply_quantity < demand_quantity * 0.5 THEN 'low'
        ELSE 'balanced'
    END;
    
    demand_level := CASE 
        WHEN demand_quantity > supply_quantity * 2 THEN 'high'
        WHEN demand_quantity < supply_quantity * 0.5 THEN 'low'
        ELSE 'balanced'
    END;
    
    -- Recomendar acción
    action := CASE 
        WHEN trend = 'rising' AND supply_level = 'low' THEN 'buy_now'
        WHEN trend = 'falling' AND supply_level = 'high' THEN 'sell_now'
        WHEN trend = 'stable' AND supply_level = 'balanced' THEN 'hold'
        WHEN trend = 'rising' AND supply_level = 'high' THEN 'wait_to_buy'
        WHEN trend = 'falling' AND supply_level = 'low' THEN 'wait_to_sell'
        ELSE 'analyze_further'
    END;
    
    RETURN QUERY
    SELECT 
        p_resource_type,
        avg_price,
        trend,
        supply_level,
        demand_level,
        action;
END;
$$ LANGUAGE plpgsql;

-- Función para procesar beneficios de alianzas
CREATE OR REPLACE FUNCTION process_alliance_benefits(p_alliance_id UUID)
RETURNS TABLE(
    member_id UUID,
    resource_bonus JSONB,
    military_bonus JSONB,
    construction_bonus JSONB,
    total_benefits JSONB
) AS $$
DECLARE
    alliance_record RECORD;
    member_record RECORD;
    resource_bonus JSONB;
    military_bonus JSONB;
    construction_bonus JSONB;
    total_benefits JSONB;
BEGIN
    -- Obtener información de la alianza
    SELECT * INTO alliance_record
    FROM alliances
    WHERE id = p_alliance_id AND is_active = true;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Alianza no encontrada o inactiva';
    END IF;
    
    -- Procesar cada miembro
    FOR member_record IN
        SELECT am.player_id, p.username
        FROM alliance_members am
        JOIN players p ON am.player_id = p.id
        WHERE am.alliance_id = p_alliance_id
    LOOP
        -- Calcular bonificación de recursos basada en nivel de alianza
        resource_bonus := jsonb_build_object(
            'wood_production', alliance_record.level * 0.02,
            'stone_production', alliance_record.level * 0.02,
            'food_production', alliance_record.level * 0.02,
            'gold_production', alliance_record.level * 0.01
        );
        
        -- Calcular bonificación militar basada en experiencia de alianza
        military_bonus := jsonb_build_object(
            'unit_attack', alliance_record.experience * 0.001,
            'unit_defense', alliance_record.experience * 0.001,
            'training_speed', alliance_record.level * 0.01
        );
        
        -- Calcular bonificación de construcción basada en nivel de alianza
        construction_bonus := jsonb_build_object(
            'construction_speed', alliance_record.level * 0.015,
            'upgrade_cost_reduction', alliance_record.level * 0.01,
            'research_speed', alliance_record.level * 0.01
        );
        
        -- Combinar todos los beneficios
        total_benefits := jsonb_build_object(
            'resource_bonus', resource_bonus,
            'military_bonus', military_bonus,
            'construction_bonus', construction_bonus,
            'alliance_level', alliance_record.level,
            'alliance_experience', alliance_record.experience
        );
        
        -- Retornar beneficios para este miembro
        member_id := member_record.player_id;
        
        RETURN NEXT;
    END LOOP;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- Función para calcular puntuación del jugador
CREATE OR REPLACE FUNCTION calculate_player_score(p_player_id UUID)
RETURNS TABLE(
    total_score BIGINT,
    level_score BIGINT,
    building_score BIGINT,
    military_score BIGINT,
    achievement_score BIGINT,
    activity_score BIGINT,
    rank_position INTEGER
) AS $$
DECLARE
    level_score BIGINT;
    building_score BIGINT;
    military_score BIGINT;
    achievement_score BIGINT;
    activity_score BIGINT;
    total_score BIGINT;
    rank_position INTEGER;
BEGIN
    -- Calcular puntuación por nivel
    SELECT 
        (p.level * 100 + p.experience)::BIGINT INTO level_score
    FROM players p
    WHERE p.id = p_player_id;
    
    -- Calcular puntuación por edificios
    SELECT 
        COALESCE(SUM(b.level * 10), 0)::BIGINT INTO building_score
    FROM buildings b
    JOIN villages v ON b.village_id = v.id
    WHERE v.player_id = p_player_id;
    
    -- Calcular puntuación militar
    SELECT 
        COALESCE(SUM(pu.quantity * ut.attack + pu.quantity * ut.defense), 0)::BIGINT INTO military_score
    FROM player_units pu
    JOIN unit_types ut ON pu.unit_type_id = ut.id
    WHERE pu.player_id = p_player_id;
    
    -- Calcular puntuación por logros
    SELECT 
        COALESCE(COUNT(*) * 50, 0)::BIGINT INTO achievement_score
    FROM player_achievements pa
    WHERE pa.player_id = p_player_id AND pa.is_completed = true;
    
    -- Calcular puntuación por actividad
    SELECT 
        CASE 
            WHEN p.last_active >= NOW() - INTERVAL '1 hour' THEN 100
            WHEN p.last_active >= NOW() - INTERVAL '24 hours' THEN 50
            WHEN p.last_active >= NOW() - INTERVAL '7 days' THEN 25
            ELSE 0
        END::BIGINT INTO activity_score
    FROM players p
    WHERE p.id = p_player_id;
    
    -- Calcular puntuación total
    total_score := level_score + building_score + military_score + achievement_score + activity_score;
    
    -- Calcular posición en ranking
    SELECT 
        COALESCE(rank_position, 0) INTO rank_position
    FROM (
        SELECT 
            player_id,
            ROW_NUMBER() OVER (ORDER BY 
                (level * 100 + experience) + 
                COALESCE(building_score, 0) + 
                COALESCE(military_score, 0) + 
                COALESCE(achievement_score, 0) + 
                activity_score DESC
            ) as rank_position
        FROM players p
        LEFT JOIN (
            SELECT 
                v.player_id,
                SUM(b.level * 10) as building_score
            FROM buildings b
            JOIN villages v ON b.village_id = v.id
            GROUP BY v.player_id
        ) bs ON p.id = bs.player_id
        LEFT JOIN (
            SELECT 
                pu.player_id,
                SUM(pu.quantity * ut.attack + pu.quantity * ut.defense) as military_score
            FROM player_units pu
            JOIN unit_types ut ON pu.unit_type_id = ut.id
            GROUP BY pu.player_id
        ) ms ON p.id = ms.player_id
        LEFT JOIN (
            SELECT 
                pa.player_id,
                COUNT(*) * 50 as achievement_score
            FROM player_achievements pa
            WHERE pa.is_completed = true
            GROUP BY pa.player_id
        ) acs ON p.id = acs.player_id
    ) ranked_players
    WHERE player_id = p_player_id;
    
    RETURN QUERY
    SELECT 
        total_score,
        level_score,
        building_score,
        military_score,
        achievement_score,
        activity_score,
        rank_position;
END;
$$ LANGUAGE plpgsql;

-- Función para generar recompensas diarias
CREATE OR REPLACE FUNCTION generate_daily_rewards(p_player_id UUID)
RETURNS TABLE(
    reward_type VARCHAR,
    reward_value INTEGER,
    reward_description TEXT,
    consecutive_days INTEGER,
    total_rewards_today INTEGER
) AS $$
DECLARE
    player_record RECORD;
    last_login_date DATE;
    consecutive_days INTEGER;
    daily_bonus INTEGER;
    vip_bonus INTEGER;
    event_bonus INTEGER;
    total_rewards INTEGER;
    reward_type VARCHAR;
    reward_value INTEGER;
    reward_description TEXT;
BEGIN
    -- Obtener información del jugador
    SELECT 
        p.*,
        COALESCE(pp.consecutive_login_days, 0) as consecutive_days
    INTO player_record
    FROM players p
    LEFT JOIN player_prestige pp ON p.id = pp.player_id
    WHERE p.id = p_player_id;
    
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Jugador no encontrado';
    END IF;
    
    -- Verificar si ya recibió recompensas hoy
    last_login_date := player_record.last_login::DATE;
    
    IF last_login_date = CURRENT_DATE THEN
        RAISE EXCEPTION 'Ya recibiste recompensas hoy';
    END IF;
    
    -- Calcular días consecutivos
    IF last_login_date = CURRENT_DATE - INTERVAL '1 day' THEN
        consecutive_days := COALESCE(player_record.consecutive_days, 0) + 1;
    ELSE
        consecutive_days := 1;
    END IF;
    
    -- Calcular bonificaciones
    daily_bonus := 10 + (consecutive_days * 2); -- Base + bonus por días consecutivos
    vip_bonus := CASE WHEN player_record.role = 'vip' THEN 20 ELSE 0 END;
    event_bonus := CASE 
        WHEN EXISTS (
            SELECT 1 FROM events e 
            WHERE e.is_active = true 
            AND e.event_type = 'daily_bonus'
            AND NOW() BETWEEN e.start_time AND e.end_time
        ) THEN 15
        ELSE 0
    END;
    
    total_rewards := daily_bonus + vip_bonus + event_bonus;
    
    -- Generar recompensa de oro
    reward_type := 'gold';
    reward_value := total_rewards;
    reward_description := format('Recompensa diaria por %s días consecutivos', consecutive_days);
    
    -- Actualizar información del jugador
    UPDATE players 
    SET 
        gold = gold + total_rewards,
        last_login = NOW(),
        updated_at = NOW()
    WHERE id = p_player_id;
    
    -- Actualizar días consecutivos
    INSERT INTO player_prestige (player_id, consecutive_login_days, last_daily_reward)
    VALUES (p_player_id, consecutive_days, NOW())
    ON CONFLICT (player_id)
    DO UPDATE SET 
        consecutive_login_days = consecutive_days,
        last_daily_reward = NOW(),
        updated_at = NOW();
    
    RETURN QUERY
    SELECT 
        reward_type,
        reward_value,
        reward_description,
        consecutive_days,
        total_rewards;
END;
$$ LANGUAGE plpgsql;

-- Función para limpieza de datos inactivos
CREATE OR REPLACE FUNCTION cleanup_inactive_data(p_days_old INTEGER DEFAULT 30)
RETURNS TABLE(
    table_name VARCHAR,
    records_deleted INTEGER,
    cleanup_type VARCHAR
) AS $$
DECLARE
    deleted_count INTEGER;
    table_name VARCHAR;
    cleanup_type VARCHAR;
BEGIN
    -- Limpiar mensajes de chat antiguos
    DELETE FROM chat_messages 
    WHERE created_at < NOW() - INTERVAL '1 day' * p_days_old;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'chat_messages';
    records_deleted := deleted_count;
    cleanup_type := 'old_messages';
    RETURN NEXT;
    
    -- Limpiar notificaciones leídas antiguas
    DELETE FROM notifications 
    WHERE is_read = true AND created_at < NOW() - INTERVAL '1 day' * p_days_old;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'notifications';
    records_deleted := deleted_count;
    cleanup_type := 'read_notifications';
    RETURN NEXT;
    
    -- Limpiar logs de rendimiento antiguos
    DELETE FROM performance_logs 
    WHERE created_at < NOW() - INTERVAL '1 day' * p_days_old;
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'performance_logs';
    records_deleted := deleted_count;
    cleanup_type := 'performance_logs';
    RETURN NEXT;
    
    -- Limpiar transacciones de moneda antiguas (mantener solo 1 año)
    DELETE FROM currency_transactions 
    WHERE created_at < NOW() - INTERVAL '1 year';
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'currency_transactions';
    records_deleted := deleted_count;
    cleanup_type := 'old_transactions';
    RETURN NEXT;
    
    -- Limpiar reportes de batalla antiguos (mantener solo 6 meses)
    DELETE FROM battle_reports 
    WHERE created_at < NOW() - INTERVAL '6 months';
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'battle_reports';
    records_deleted := deleted_count;
    cleanup_type := 'old_battle_reports';
    RETURN NEXT;
    
    -- Limpiar datos temporales de eventos
    DELETE FROM event_participation 
    WHERE event_id IN (
        SELECT id FROM events 
        WHERE end_time < NOW() - INTERVAL '1 day' * p_days_old
    );
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    table_name := 'event_participation';
    records_deleted := deleted_count;
    cleanup_type := 'old_event_data';
    RETURN NEXT;
    
    RETURN;
END;
$$ LANGUAGE plpgsql;

-- ========================================
-- TRIGGERS
-- ========================================

-- Trigger para actualizar timestamp de actualización en players
CREATE TRIGGER update_players_updated_at
    BEFORE UPDATE ON players
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para actualizar timestamp de actualización en alliances
CREATE TRIGGER update_alliances_updated_at
    BEFORE UPDATE ON alliances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para actualizar timestamp de actualización en worlds
CREATE TRIGGER update_worlds_updated_at
    BEFORE UPDATE ON worlds
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para actualizar timestamp de actualización en player_technologies
CREATE TRIGGER update_player_technologies_updated_at
    BEFORE UPDATE ON player_technologies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para actualizar timestamp de actualización en player_heroes
CREATE TRIGGER update_player_heroes_updated_at
    BEFORE UPDATE ON player_heroes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Trigger para actualizar timestamp de actualización en player_units
CREATE TRIGGER update_player_units_updated_at
    BEFORE UPDATE ON player_units
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ========================================
-- VISTAS ÚTILES
-- ========================================

-- Vista de estadísticas del servidor
CREATE OR REPLACE VIEW server_statistics AS
SELECT 
    (SELECT COUNT(*) FROM players WHERE is_online = true) as active_players,
    (SELECT COUNT(*) FROM players WHERE is_online = true) as online_players,
    (SELECT COUNT(*) FROM villages) as total_villages,
    (SELECT COUNT(*) FROM alliances WHERE is_active = true) as active_alliances,
    (SELECT COUNT(*) FROM battles WHERE status = 'in_progress') as active_battles,
    (SELECT COUNT(*) FROM chat_messages WHERE created_at >= NOW() - INTERVAL '1 hour') as messages_last_hour;

-- Vista de aldeas con detalles
CREATE OR REPLACE VIEW village_details AS
SELECT 
    v.id,
    v.name,
    v.x_coordinate,
    v.y_coordinate,
    v.created_at,
    p.username as player_name,
    w.name as world_name,
    r.wood,
    r.stone,
    r.food,
    r.gold,
    r.last_updated
FROM villages v
JOIN players p ON v.player_id = p.id
JOIN worlds w ON v.world_id = w.id
LEFT JOIN resources r ON r.village_id = v.id;

-- Vista de jugadores con estadísticas
CREATE OR REPLACE VIEW player_statistics AS
SELECT 
    p.id,
    p.username,
    p.level,
    p.experience,
    p.gold,
    p.gems,
    p.is_online,
    p.last_active,
    COUNT(v.id) as village_count,
    COALESCE(SUM(pu.quantity), 0) as total_units,
    COUNT(DISTINCT pa.achievement_id) as achievements_completed
FROM players p
LEFT JOIN villages v ON p.id = v.player_id
LEFT JOIN player_units pu ON p.id = pu.player_id
LEFT JOIN player_achievements pa ON p.id = pa.player_id AND pa.is_completed = true
GROUP BY p.id, p.username, p.level, p.experience, p.gold, p.gems, p.is_online, p.last_active;

-- ========================================
-- PERMISOS Y ROLES
-- ========================================

-- Crear usuario de aplicación (ajustar según tu configuración)
-- CREATE USER etheria_user WITH PASSWORD 'tu_password_aqui';
-- GRANT ALL PRIVILEGES ON DATABASE tu_base_de_datos TO etheria_user;
-- GRANT ALL PRIVILEGES ON SCHEMA public TO etheria_user;
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO etheria_user;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO etheria_user;

-- ========================================
-- MENSAJE DE COMPLETADO
-- ========================================

DO $$
BEGIN
    RAISE NOTICE 'Base de datos MMO creada exitosamente!';
    RAISE NOTICE 'Tablas creadas: %', (SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public');
    RAISE NOTICE 'Índices creados: %', (SELECT COUNT(*) FROM pg_indexes WHERE schemaname = 'public');
    RAISE NOTICE 'Funciones creadas: %', (SELECT COUNT(*) FROM information_schema.routines WHERE routine_schema = 'public');
    RAISE NOTICE 'Vistas creadas: %', (SELECT COUNT(*) FROM information_schema.views WHERE table_schema = 'public');
END $$;

-- ========================================
-- SISTEMA DE TÍTULOS Y PRESTIGIO
-- ========================================

-- Tabla de categorías de títulos
CREATE TABLE IF NOT EXISTS title_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    color VARCHAR(7),
    background_color VARCHAR(7),
    display_order INTEGER DEFAULT 0,
    is_public BOOLEAN DEFAULT true,
    show_in_profile BOOLEAN DEFAULT true,
    show_in_dashboard BOOLEAN DEFAULT true,
    total_titles INTEGER DEFAULT 0,
    unlocked_titles INTEGER DEFAULT 0,
    unlocked_count INTEGER DEFAULT 0,
    total_players INTEGER DEFAULT 0,
    completion_rate DECIMAL(5,2) DEFAULT 0.0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de títulos
CREATE TABLE IF NOT EXISTS titles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID REFERENCES title_categories(id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    long_description TEXT,
    story_text TEXT,
    icon VARCHAR(100),
    color VARCHAR(7),
    background_color VARCHAR(7),
    border_color VARCHAR(7),
    rarity VARCHAR(20) DEFAULT 'common',
    title_type VARCHAR(50) NOT NULL,
    title_format VARCHAR(20),
    display_format VARCHAR(100),
    level_required INTEGER DEFAULT 1,
    prestige_required INTEGER DEFAULT 0,
    alliance_required UUID,
    prerequisites TEXT,
    unlock_conditions TEXT,
    effects TEXT,
    bonuses TEXT,
    special_abilities TEXT,
    prestige_value INTEGER DEFAULT 0,
    reputation_bonus INTEGER DEFAULT 0,
    social_status VARCHAR(50),
    max_owners INTEGER DEFAULT 0,
    time_limit INTEGER DEFAULT 0,
    is_exclusive BOOLEAN DEFAULT false,
    is_temporary BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'available',
    unlock_date TIMESTAMP WITH TIME ZONE,
    retire_date TIMESTAMP WITH TIME ZONE,
    total_unlocked INTEGER DEFAULT 0,
    current_owners INTEGER DEFAULT 0,
    unlock_rate DECIMAL(5,2) DEFAULT 0.0,
    is_repeatable BOOLEAN DEFAULT false,
    repeat_interval VARCHAR(20),
    next_unlock_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    is_hidden BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de títulos de jugadores
CREATE TABLE IF NOT EXISTS player_titles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'active',
    is_unlocked BOOLEAN DEFAULT false,
    is_equipped BOOLEAN DEFAULT false,
    is_favorite BOOLEAN DEFAULT false,
    unlock_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    equipped_date TIMESTAMP WITH TIME ZONE,
    unlock_method VARCHAR(50),
    unlock_data TEXT,
    expiry_date TIMESTAMP WITH TIME ZONE,
    days_remaining INTEGER DEFAULT 0,
    is_permanent BOOLEAN DEFAULT true,
    times_equipped INTEGER DEFAULT 0,
    total_time_equipped INTEGER DEFAULT 0,
    last_equipped TIMESTAMP WITH TIME ZONE,
    progress INTEGER DEFAULT 0,
    max_progress INTEGER DEFAULT 100,
    level INTEGER DEFAULT 1,
    max_level INTEGER DEFAULT 1,
    prestige_level INTEGER DEFAULT 0,
    rewards_claimed BOOLEAN DEFAULT false,
    rewards_data TEXT,
    points_earned INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(player_id, title_id)
);

-- Tabla de niveles de prestigio
CREATE TABLE IF NOT EXISTS prestige_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    color VARCHAR(7),
    background_color VARCHAR(7),
    level INTEGER NOT NULL,
    prestige_required INTEGER NOT NULL,
    experience_multiplier DECIMAL(5,2) DEFAULT 1.0,
    bonuses TEXT,
    special_effects TEXT,
    unlock_features TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de prestigio de jugadores
CREATE TABLE IF NOT EXISTS player_prestige (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    current_prestige INTEGER DEFAULT 0,
    total_prestige INTEGER DEFAULT 0,
    prestige_level INTEGER DEFAULT 1,
    prestige_to_next INTEGER DEFAULT 0,
    progress_percent DECIMAL(5,2) DEFAULT 0.0,
    titles_unlocked INTEGER DEFAULT 0,
    titles_equipped INTEGER DEFAULT 0,
    achievements_completed INTEGER DEFAULT 0,
    prestige_history TEXT,
    last_gain_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    largest_gain INTEGER DEFAULT 0,
    global_rank INTEGER DEFAULT 0,
    category_rank INTEGER DEFAULT 0,
    alliance_rank INTEGER DEFAULT 0,
    first_prestige TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de logros de títulos
CREATE TABLE IF NOT EXISTS title_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    achievement_type VARCHAR(50) NOT NULL,
    requirements TEXT,
    progress_type VARCHAR(20),
    target_value INTEGER DEFAULT 1,
    rewards TEXT,
    prestige_reward INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de logros de títulos de jugadores
CREATE TABLE IF NOT EXISTS player_title_achievements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    achievement_id UUID NOT NULL REFERENCES title_achievements(id) ON DELETE CASCADE,
    current_progress INTEGER DEFAULT 0,
    is_completed BOOLEAN DEFAULT false,
    completion_date TIMESTAMP WITH TIME ZONE,
    rewards_claimed BOOLEAN DEFAULT false,
    claim_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(player_id, achievement_id)
);

-- Tabla de eventos de títulos
CREATE TABLE IF NOT EXISTS title_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(100),
    event_type VARCHAR(50) NOT NULL,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    effects TEXT,
    prestige_multiplier DECIMAL(5,2) DEFAULT 1.0,
    unlock_chance DECIMAL(5,2) DEFAULT 0.0,
    total_participants INTEGER DEFAULT 0,
    active_participants INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'upcoming',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de rankings de títulos
CREATE TABLE IF NOT EXISTS title_rankings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    prestige_rank INTEGER DEFAULT 0,
    titles_rank INTEGER DEFAULT 0,
    achievements_rank INTEGER DEFAULT 0,
    overall_rank INTEGER DEFAULT 0,
    prestige_score INTEGER DEFAULT 0,
    titles_score INTEGER DEFAULT 0,
    achievements_score INTEGER DEFAULT 0,
    overall_score INTEGER DEFAULT 0,
    rare_titles INTEGER DEFAULT 0,
    epic_titles INTEGER DEFAULT 0,
    legendary_titles INTEGER DEFAULT 0,
    mythic_titles INTEGER DEFAULT 0,
    divine_titles INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de estadísticas de títulos
CREATE TABLE IF NOT EXISTS title_statistics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    total_titles_unlocked INTEGER DEFAULT 0,
    total_titles_equipped INTEGER DEFAULT 0,
    total_prestige_gained INTEGER DEFAULT 0,
    total_achievements_completed INTEGER DEFAULT 0,
    category_stats TEXT,
    type_stats TEXT,
    rarity_stats TEXT,
    last_title_unlocked TIMESTAMP WITH TIME ZONE,
    last_title_equipped TIMESTAMP WITH TIME ZONE,
    last_prestige_gain TIMESTAMP WITH TIME ZONE,
    longest_equipped_title VARCHAR(100),
    most_prestigious_title VARCHAR(100),
    rarest_title VARCHAR(100),
    first_title_unlocked TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de notificaciones de títulos
CREATE TABLE IF NOT EXISTS title_notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    player_id UUID NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    title_id UUID REFERENCES titles(id),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    data TEXT,
    is_read BOOLEAN DEFAULT false,
    is_dismissed BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    read_at TIMESTAMP WITH TIME ZONE,
    dismissed_at TIMESTAMP WITH TIME ZONE
);

-- Tabla de recompensas de títulos
CREATE TABLE IF NOT EXISTS title_rewards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title_id UUID NOT NULL REFERENCES titles(id) ON DELETE CASCADE,
    reward_type VARCHAR(50) NOT NULL,
    reward_data TEXT,
    quantity INTEGER DEFAULT 1,
    level_required INTEGER DEFAULT 0,
    is_repeatable BOOLEAN DEFAULT false,
    is_guaranteed BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Tabla de leaderboards de títulos
CREATE TABLE IF NOT EXISTS title_leaderboards (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    category_id UUID REFERENCES title_categories(id),
    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    total_participants INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
