package server

import (
	"net"

	pb "github.com/bllooop/pvzservice/grpcpvz"
	"github.com/bllooop/pvzservice/internal/delivery/api"
	"github.com/bllooop/pvzservice/internal/usecase"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"google.golang.org/grpc"
)

func StartGRPC(port string, usecase *usecase.Usecase) *grpc.Server {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		logger.Log.Fatal().Msg("При запуске GRPC сервера произошла ошибка")
	}
	grpcServer := grpc.NewServer()
	pbzSrv := api.NewPVZServiceServer(usecase)
	pb.RegisterPVZServiceServer(grpcServer, pbzSrv)
	logger.Log.Info().Msgf("Сервер работает на порту %v", lis.Addr())
	go func() {
		if err := grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Log.Error().Err(err).Msg("Ошибка при работе gRPC сервера")
		}
	}()
	return grpcServer
}
