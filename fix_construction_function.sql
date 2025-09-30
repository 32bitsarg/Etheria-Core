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
            upgrade_completion_time = NULL
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


