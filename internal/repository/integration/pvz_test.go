package integration

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/bllooop/pvzservice/internal/domain"
	"github.com/bllooop/pvzservice/internal/repository"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PvzRepoTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *PostgresContainer
	repository  *repository.PvzPostgres
	db          *sqlx.DB
}

func (suite *PvzRepoTestSuite) SetupSuite() {
	suite.ctx = context.Background()
	pgContainer, err := CreatePostgresContainer(suite.ctx)
	if err != nil {
		logger.Log.Fatal().Err(err)
	}
	host, err := pgContainer.Host(suite.ctx)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to get container host")
	}
	port, err := pgContainer.MappedPort(suite.ctx, "5432")
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to get mapped port")
	}

	cfg := repository.Config{
		Username: "postgres",
		Password: "postgres",
		Host:     host,
		Port:     port.Port(),
		DBname:   "test-db",
		SSLMode:  "disable",
	}
	suite.pgContainer = pgContainer
	db, err := repository.NewPostgresDB(cfg)
	if err != nil {
		logger.Log.Fatal().Err(err)
	}
	suite.db = db
	migratePath, err := filepath.Abs("../../../migrations")
	fmt.Println("DEBUG: Running migrations from", migratePath)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("Migration failed")
	}
	err = repository.RunMigrate(cfg, migratePath)

	if err != nil {
		logger.Log.Fatal().Err(err)

	}
	suite.repository = repository.NewPvzPostgres(suite.db)

}
func (suite *PvzRepoTestSuite) SetupTest() {
	_, err := suite.repository.DB().Exec("TRUNCATE TABLE userlist, pvz, product_reception, product RESTART IDENTITY CASCADE")
	assert.NoError(suite.T(), err)
}
func (suite *PvzRepoTestSuite) TearDownSuite() {
	if err := suite.pgContainer.Terminate(suite.ctx); err != nil {
		logger.Log.Fatal().Err(err).Msg("error terminating postgres container")
	}
}

func (suite *PvzRepoTestSuite) TestPvzWork() {
	fixedTime := time.Date(2025, 4, 10, 15, 5, 17, 329922000, time.UTC)
	pvzIDTest, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}
	stat := "in_progress"
	statClose := "close"
	typeProd := "электроника"
	t := suite.T()
	//_, err = suite.repository.DB().Exec("INSERT INTO pvzTable (registrationDate,city) VALUES ($1,$2) ON CONFLICT (id) DO NOTHING", "Москва", fixedTime)
	//assert.NoError(t, err)
	input := domain.PVZ{
		DateRegister: &fixedTime,
		City:         "Москва",
	}

	createdPvz, err := suite.repository.CreatePvz(input)
	if err != nil {
		t.Fatalf("Failed to createPvz: %s", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, createdPvz)
	assert.NotNil(t, createdPvz.Id)
	assert.Equal(t, "Москва", createdPvz.City)

	pvzIDTest = *createdPvz.Id

	inputRecep := domain.ProductReception{
		DateReceived: &fixedTime,
		Status:       &stat,
		PVZId:        &pvzIDTest,
	}
	createdRecep, err := suite.repository.CreateRecep(inputRecep)
	if err != nil {
		t.Fatalf("Failed to createRecep: %s", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, createdRecep)
	assert.NotNil(t, createdRecep.Id)
	assert.Equal(t, pvzIDTest, *createdRecep.PVZId)
	assert.Equal(t, stat, *createdRecep.Status)

	receptionID := *createdRecep.Id
	for i := 0; i < 50; i++ {
		inputProd := domain.Product{
			DateReceived: &fixedTime,
			PVZId:        &pvzIDTest,
			Type:         typeProd,
		}

		addedProd, err := suite.repository.AddProdToRecep(inputProd)
		if err != nil {
			t.Fatalf("Failed to add Product: %s", err)
		}
		assert.NoError(t, err)
		assert.NotNil(t, addedProd)
		assert.NotNil(t, addedProd.Id)
		assert.Equal(t, receptionID, *addedProd.ReceptionId)
		assert.Equal(t, typeProd, addedProd.Type)
	}
	closedRecep, err := suite.repository.CloseReception(*inputRecep.PVZId)
	if err != nil {
		t.Fatalf("Failed to close Reception: %s", err)
	}
	assert.NoError(t, err)
	assert.NotNil(t, closedRecep)
	assert.Equal(t, statClose, *closedRecep.Status)
	assert.Equal(t, receptionID, *closedRecep.Id)
}

func TestCustomerRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PvzRepoTestSuite))
}

func IntPointer(s int) *int {
	return &s
}
