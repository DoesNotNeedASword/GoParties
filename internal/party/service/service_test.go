package service

import (
	"context"
	"errors"
	"testing"

	"Parties/internal/party/model"
)

func TestCreateTrimsInputAndAllowsParty(t *testing.T) {
	repository := &fakeRepository{}
	service := New(repository)

	party, err := service.Create(context.Background(), model.CreatePartyInput{
		Name:        "  Party  ",
		Description: "  Description  ",
		Promises:    "  Promises  ",
		Image:       "  image.png  ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if party.Name != "Party" {
		t.Fatalf("Name = %q, want %q", party.Name, "Party")
	}
	if party.Description != "Description" {
		t.Fatalf("Description = %q, want %q", party.Description, "Description")
	}
	if party.Promises != "Promises" {
		t.Fatalf("Promises = %q, want %q", party.Promises, "Promises")
	}
	if party.Image != "image.png" {
		t.Fatalf("Image = %q, want %q", party.Image, "image.png")
	}
	if !party.IsAllowed {
		t.Fatal("IsAllowed = false, want true")
	}
}

func TestCreateRejectsInvalidInput(t *testing.T) {
	service := New(&fakeRepository{})

	_, err := service.Create(context.Background(), model.CreatePartyInput{
		Name:        "Party",
		Description: "Description",
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestListNormalizesFilter(t *testing.T) {
	repository := &fakeRepository{}
	service := New(repository)

	_, err := service.List(context.Background(), model.ListPartiesFilter{
		Limit:  1000,
		Offset: -5,
	})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if repository.lastFilter.Limit != 100 {
		t.Fatalf("Limit = %d, want %d", repository.lastFilter.Limit, 100)
	}
	if repository.lastFilter.Offset != 0 {
		t.Fatalf("Offset = %d, want %d", repository.lastFilter.Offset, 0)
	}
}

func TestBanSetsReason(t *testing.T) {
	reason := "illegal funding"
	repository := &fakeRepository{
		party: model.Party{
			ID:          10,
			Name:        "Party",
			Description: "Description",
			Promises:    "Promises",
			Image:       "image.png",
			IsAllowed:   true,
		},
	}
	service := New(repository)

	party, err := service.Ban(context.Background(), 10, "  "+reason+"  ")
	if err != nil {
		t.Fatalf("Ban() error = %v", err)
	}

	if party.IsAllowed {
		t.Fatal("IsAllowed = true, want false")
	}
	if party.BanReason == nil || *party.BanReason != reason {
		t.Fatalf("BanReason = %v, want %q", party.BanReason, reason)
	}
}

func TestUpdateRejectsBannedPartyWithoutReason(t *testing.T) {
	service := New(&fakeRepository{})

	_, err := service.Update(context.Background(), model.UpdatePartyInput{
		ID:          10,
		Name:        "Party",
		Description: "Description",
		Promises:    "Promises",
		IsAllowed:   false,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("Update() error = %v, want %v", err, ErrInvalidInput)
	}
}

func TestAllowClearsReason(t *testing.T) {
	reason := "illegal funding"
	repository := &fakeRepository{
		party: model.Party{
			ID:          10,
			Name:        "Party",
			Description: "Description",
			Promises:    "Promises",
			Image:       "image.png",
			IsAllowed:   false,
			BanReason:   &reason,
		},
	}
	service := New(repository)

	party, err := service.Allow(context.Background(), 10)
	if err != nil {
		t.Fatalf("Allow() error = %v", err)
	}

	if !party.IsAllowed {
		t.Fatal("IsAllowed = false, want true")
	}
	if party.BanReason != nil {
		t.Fatalf("BanReason = %v, want nil", *party.BanReason)
	}
}

type fakeRepository struct {
	party      model.Party
	lastFilter model.ListPartiesFilter
}

func (r *fakeRepository) Create(ctx context.Context, party model.Party) (model.Party, error) {
	_ = ctx
	party.ID = 1
	r.party = party
	return party, nil
}

func (r *fakeRepository) GetByID(ctx context.Context, id int64) (model.Party, error) {
	_ = ctx
	if id != r.party.ID {
		return model.Party{}, errors.New("not found")
	}
	return r.party, nil
}

func (r *fakeRepository) List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error) {
	_ = ctx
	r.lastFilter = filter
	return []model.Party{r.party}, nil
}

func (r *fakeRepository) Update(ctx context.Context, party model.Party) (model.Party, error) {
	_ = ctx
	r.party = party
	return party, nil
}

func (r *fakeRepository) Delete(ctx context.Context, id int64) error {
	_ = ctx
	_ = id
	return nil
}
