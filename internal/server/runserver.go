package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handlers "github.com/bllooop/pvzservice/internal/delivery/api"
	"github.com/bllooop/pvzservice/internal/repository"
	"github.com/bllooop/pvzservice/internal/usecase"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Run() {
	logger.Log.Debug().Msg("Инициализация сервера...")

	if err := initConfig(); err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("Возникла ошибка загрузки конфига")
	}
	if err := godotenv.Load(); err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("Возникла ошибка с env")
	}
	logger.Log.Debug().Msg("Переменные окружения успешно загружены")
	dbpool, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBname:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})
	if err != nil {
		logger.Log.Error().Err(err).Msg("Не удалось установить соединение с базой данных")
		logger.Log.Fatal().Msg("Произошла ошибка с базой данных")
	}
	logger.Log.Debug().Msg("База данных успешно подключена")

	migratePath := "./migrations"
	logger.Log.Debug().Msgf("Running database migrations from path: %s", migratePath)
	if err = repository.RunMigrate(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBname:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	}, migratePath); err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("Возникла ошибка при переносе")
	}
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("Произошла ошибка с базой данных")
	}
	logger.Log.Debug().Msg("Инициализация слоя репозитория")
	repos := repository.NewRepository(dbpool)
	logger.Log.Debug().Msg("Инициализация usecase слоя")
	usecases := usecase.NewUsecase(repos)
	logger.Log.Debug().Msg("Инициализация обработчиков API")
	handler := handlers.NewHandler(usecases)
	srv := new(Server)
	//http serv
	go func() {
		logger.Log.Info().Msg("Запуск сервера...")
		if err := srv.StartHTTP(viper.GetString("port"), handler.InitRoutes()); err != nil && err != http.ErrServerClosed {
			logger.Log.Error().Err(err).Msg("")
			logger.Log.Fatal().Msg("При запуске HTTP сервера произошла ошибка")
		} else {
			logger.Log.Info().Msg("HTTP сервер был закрыт аккуратно")
		}
	}()
	//grpc serv
	logger.Log.Info().Msg("Запуск сервера gRPC...")
	grpcServer := StartGRPC(viper.GetString("portGrpc"), usecases)
	logger.Log.Info().Msg("Сервер HTTP и gRPC работает")
	go func() {
		logger.Log.Info().Msg("Попытка подключения к клиенту GRPC")
		err := CallGRPCClient()
		if err != nil {
			logger.Log.Error().Err(err).Msg("Ошибка подключения к клиенту gRPC")
		}

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				logger.Log.Info().Msg("Вызов клиента gRPC")
				err := CallGRPCClient()
				if err != nil {
					logger.Log.Error().Err(err).Msg("Вызов gRPC закончился с ошибкой")
					logger.Log.Fatal().Msg("Ошибка подключения")
				} else {
					logger.Log.Info().Msg("Успешно вызван клиент gRPC")
				}
			}
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	logger.Log.Debug().Msg("Прослушивание сигналов завершения работы ОС")
	<-quit
	logger.Log.Info().Msg("Сервер отключается")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer dbpool.Close()
	logger.Log.Debug().Msg("Закрытие соединения с базой данных ")
	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("При выключении сервера произошла ошибка")
	}
	grpcServer.GracefulStop()
	logger.Log.Info().Msg("gRPC сервер отключен")
}

func initConfig() error {
	viper.AddConfigPath("./config")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}
