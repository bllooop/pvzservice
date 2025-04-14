package api

import (
	"context"

	pb "github.com/bllooop/pvzservice/grpcpvz"
	"github.com/bllooop/pvzservice/internal/usecase"
	logger "github.com/bllooop/pvzservice/pkg/logging"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServiceServerHandle struct {
	usecase *usecase.Usecase
	pb.UnimplementedPVZServiceServer
}

func NewPVZServiceServer(s *usecase.Usecase) *PVZServiceServerHandle {
	return &PVZServiceServerHandle{usecase: s}
}
func (g *PVZServiceServerHandle) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	pvzs, err := g.usecase.GetListOFpvz(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("")
		return nil, err
	}

	var pvzList []*pb.PVZ
	for _, pvz := range pvzs {
		var registrationDate *timestamppb.Timestamp
		if pvz.DateRegister != nil {
			registrationDate = timestamppb.New(*pvz.DateRegister)
		}
		pvzList = append(pvzList, &pb.PVZ{
			Id:               pvz.Id.String(),
			RegistrationDate: registrationDate,
			City:             pvz.City,
		})
	}
	logger.Log.Debug().Any("pvz", pvzList).Msg("Получен список ПВЗ")

	return &pb.GetPVZListResponse{Pvzs: pvzList}, nil
}
