package integration

import (
	"testing"

	"github.com/bllooop/pvzservice/internal/repository"
	"github.com/stretchr/testify/assert"
)

func TestNewPostgresDB(t *testing.T) {
	cfg := repository.Config{
		Host:     "localhost",
		Port:     "5436",
		Username: "postgres",
		Password: "54321",
		DBname:   "test-db",
		SSLMode:  "disable",
	}

	tempCfg := cfg
	tempCfg.DBname = "postgres"
	db, err := repository.NewPostgresDB(tempCfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	var exists bool
	err = db.QueryRow(`SELECT EXISTS (
        SELECT 1 FROM pg_database WHERE datname = 'test-db'
    )`).Scan(&exists)
	assert.NoError(t, err)

	if !exists {
		_, err = db.Exec(`CREATE DATABASE "test-db"`)
		assert.NoError(t, err)
	}

	tempCfg.DBname = "test-db"
	db, err = repository.NewPostgresDB(tempCfg)
	assert.NoError(t, err)
	assert.NotNil(t, db)

	err = db.Ping()
	assert.NoError(t, err)
}
