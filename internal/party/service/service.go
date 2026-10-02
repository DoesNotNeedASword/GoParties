package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"Parties/internal/party/model"
)

var ErrInvalidInput = errors.New("invalid party input")

type Repository interface {
	Create(ctx context.Context, party model.Party) (model.Party, error)
	GetByID(ctx context.Context, id int64) (model.Party, error)
	List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error)
	Update(ctx context.Context, party model.Party) (model.Party, error)
	Delete(ctx context.Context, id int64) error
}

type Service struct {
	repository Repository
	logger     *slog.Logger
}

func New(repository Repository, logs ...*slog.Logger) *Service {
	var log *slog.Logger
	if len(logs) > 0 {
		log = logs[0]
	}

	return &Service{repository: repository, logger: log}
}

func (s *Service) Create(ctx context.Context, input model.CreatePartyInput) (model.Party, error) {
	party := model.Party{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Promises:    strings.TrimSpace(input.Promises),
		Image:       strings.TrimSpace(input.Image),
		IsAllowed:   true,
	}

	if !isPartyValid(party) {
		return model.Party{}, ErrInvalidInput
	}

	return s.repository.Create(ctx, party)
}

func (s *Service) GetByID(ctx context.Context, id int64) (model.Party, error) {
	if id <= 0 {
		return model.Party{}, ErrInvalidInput
	}

	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	return s.repository.List(ctx, filter)
}

func (s *Service) Update(ctx context.Context, input model.UpdatePartyInput) (model.Party, error) {
	party := model.Party{
		ID:          input.ID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Promises:    strings.TrimSpace(input.Promises),
		Image:       strings.TrimSpace(input.Image),
		IsAllowed:   input.IsAllowed,
		BanReason:   trimOptional(input.BanReason),
	}

	if party.ID <= 0 || !isPartyValid(party) {
		return model.Party{}, ErrInvalidInput
	}
	if party.IsAllowed {
		party.BanReason = nil
	} else if party.BanReason == nil {
		return model.Party{}, ErrInvalidInput
	}

	return s.repository.Update(ctx, party)
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidInput
	}

	return s.repository.Delete(ctx, id)
}

func (s *Service) Ban(ctx context.Context, id int64, reason string) (model.Party, error) {
	reason = strings.TrimSpace(reason)
	if id <= 0 || reason == "" {
		return model.Party{}, ErrInvalidInput
	}

	party, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return model.Party{}, err
	}

	party.IsAllowed = false
	party.BanReason = &reason

	return s.repository.Update(ctx, party)
}

func (s *Service) Allow(ctx context.Context, id int64) (model.Party, error) {
	if id <= 0 {
		return model.Party{}, ErrInvalidInput
	}

	party, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return model.Party{}, err
	}

	party.IsAllowed = true
	party.BanReason = nil

	return s.repository.Update(ctx, party)
}

func isPartyValid(party model.Party) bool {
	return party.Name != "" && party.Description != "" && party.Promises != ""
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
