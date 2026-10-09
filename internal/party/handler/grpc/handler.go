package grpc

import (
	"Parties/internal/party/service"
	"context"

	partyv1 "github.com/DoesNotNeedASword/proto-contracts/gen/go/party/v1"
)

type PartyGRPCHandler struct {
	partyv1.UnimplementedPartyServiceServer
	service *service.Service
}

func New(service *service.Service) *PartyGRPCHandler {
	return &PartyGRPCHandler{service: service}
}

func (h *PartyGRPCHandler) GetParty(ctx context.Context, req *partyv1.GetPartyRequest) (*partyv1.GetPartyResponse, error) {
	party, err := h.service.GetByID(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &partyv1.GetPartyResponse{
		Party: &partyv1.Party{
			Id:          party.ID,
			Name:        party.Name,
			Description: party.Description,
			Promises:    party.Promises,
			ImageUrl:    party.Image,
		},
	}, nil
}
