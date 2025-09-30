package config

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func GetDBConnection() (*sql.DB, error) {
	// Cargar configuración usando la función existente
	config, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración: %v", err)
	}

	// Construir string de conexión
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.DBName,
		config.Database.SSLMode,
	)

	// Conectar a la base de datos
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error conectando a la base de datos: %v", err)
	}

	// Verificar conexión
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error verificando conexión a la base de datos: %v", err)
	}

	return db, nil
}
