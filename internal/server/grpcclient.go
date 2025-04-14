package server

import (
	"context"
	"time"

	pb "github.com/bllooop/pvzservice/grpcpvz"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func CallGRPCClient() error {
	conn, err := grpc.NewClient("pvzservice"+viper.GetString("portGrpc"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка подключения")
		return err
	}
	defer conn.Close()
	logger.Log.Info().Msg("gRPC клиент успешно подключен")
	client := pb.NewPVZServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	response, err := client.GetPVZList(ctx, &pb.GetPVZListRequest{})
	if err != nil {
		logger.Log.Error().Err(err).Msg("Ошибка вызова GetPvzList")
		return err
	}
	logger.Log.Info().Msgf("Получен список PVZ в количестве %d", len(response.Pvzs))
	for _, pvz := range response.Pvzs {
		logger.Log.Info().Msgf("ID ПВЗ: %s, Город: %s", pvz.Id, pvz.City)
		if pvz.RegistrationDate != nil {
			logger.Log.Info().Msgf("Дата регистрации ПВЗ: %v", pvz.RegistrationDate.AsTime())
		}
	}
	return nil
}
