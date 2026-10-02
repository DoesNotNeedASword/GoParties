package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"testing"

	"Parties/internal/party/model"
	"Parties/internal/party/repository"

	"github.com/go-chi/chi/v5"
)

func TestCreate(t *testing.T) {
	service := &fakeService{}
	router := newTestRouter(service)

	requestBody := bytes.NewBufferString(`{
		"name": "Party",
		"description": "Description",
		"promises": "Promises",
		"image": "image.png"
	}`)
	request := httptest.NewRequest(nethttp.MethodPost, "/parties", requestBody)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, nethttp.StatusCreated, response.Body.String())
	}
	if service.created.Name != "Party" {
		t.Fatalf("created name = %q, want %q", service.created.Name, "Party")
	}

	var body partyResponse
	decodeResponse(t, response, &body)
	if body.ID != 1 {
		t.Fatalf("id = %d, want %d", body.ID, 1)
	}
}

func TestList(t *testing.T) {
	service := &fakeService{
		parties: []model.Party{
			{
				ID:          1,
				Name:        "Party",
				Description: "Description",
				Promises:    "Promises",
				Image:       "image.png",
				IsAllowed:   true,
			},
		},
	}
	router := newTestRouter(service)
	request := httptest.NewRequest(nethttp.MethodGet, "/parties?limit=10&offset=5", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, nethttp.StatusOK, response.Body.String())
	}
	if service.lastFilter.Limit != 10 {
		t.Fatalf("limit = %d, want %d", service.lastFilter.Limit, 10)
	}
	if service.lastFilter.Offset != 5 {
		t.Fatalf("offset = %d, want %d", service.lastFilter.Offset, 5)
	}

	var body []partyResponse
	decodeResponse(t, response, &body)
	if len(body) != 1 {
		t.Fatalf("response length = %d, want %d", len(body), 1)
	}
}

func TestGetByIDMapsNotFound(t *testing.T) {
	service := &fakeService{err: repository.ErrNotFound}
	router := newTestRouter(service)
	request := httptest.NewRequest(nethttp.MethodGet, "/parties/123", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusNotFound {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, nethttp.StatusNotFound, response.Body.String())
	}
}

func TestGetByIDRejectsInvalidID(t *testing.T) {
	service := &fakeService{}
	router := newTestRouter(service)
	request := httptest.NewRequest(nethttp.MethodGet, "/parties/not-number", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, nethttp.StatusBadRequest, response.Body.String())
	}
}

func TestDelete(t *testing.T) {
	service := &fakeService{}
	router := newTestRouter(service)
	request := httptest.NewRequest(nethttp.MethodDelete, "/parties/42", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusNoContent {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, nethttp.StatusNoContent, response.Body.String())
	}
	if service.deletedID != 42 {
		t.Fatalf("deleted id = %d, want %d", service.deletedID, 42)
	}
}

func newTestRouter(service Service) nethttp.Handler {
	router := chi.NewRouter()
	New(service).RegisterRoutes(router)
	return router
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, target any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

type fakeService struct {
	party      model.Party
	parties    []model.Party
	created    model.CreatePartyInput
	updated    model.UpdatePartyInput
	lastFilter model.ListPartiesFilter
	deletedID  int64
	err        error
}

func (s *fakeService) Create(ctx context.Context, input model.CreatePartyInput) (model.Party, error) {
	_ = ctx
	if s.err != nil {
		return model.Party{}, s.err
	}
	s.created = input
	return model.Party{
		ID:          1,
		Name:        input.Name,
		Description: input.Description,
		Promises:    input.Promises,
		Image:       input.Image,
		IsAllowed:   true,
	}, nil
}

func (s *fakeService) GetByID(ctx context.Context, id int64) (model.Party, error) {
	_ = ctx
	if s.err != nil {
		return model.Party{}, s.err
	}
	if s.party.ID == 0 {
		return model.Party{ID: id}, nil
	}
	return s.party, nil
}

func (s *fakeService) List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error) {
	_ = ctx
	if s.err != nil {
		return nil, s.err
	}
	s.lastFilter = filter
	return s.parties, nil
}

func (s *fakeService) Update(ctx context.Context, input model.UpdatePartyInput) (model.Party, error) {
	_ = ctx
	if s.err != nil {
		return model.Party{}, s.err
	}
	s.updated = input
	return model.Party{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Promises:    input.Promises,
		Image:       input.Image,
		IsAllowed:   input.IsAllowed,
		BanReason:   input.BanReason,
	}, nil
}

func (s *fakeService) Delete(ctx context.Context, id int64) error {
	_ = ctx
	if s.err != nil {
		return s.err
	}
	s.deletedID = id
	return nil
}

func (s *fakeService) Ban(ctx context.Context, id int64, reason string) (model.Party, error) {
	_ = ctx
	if s.err != nil {
		return model.Party{}, s.err
	}
	if reason == "" {
		return model.Party{}, errors.New("empty reason")
	}
	return model.Party{
		ID:        id,
		IsAllowed: false,
		BanReason: &reason,
	}, nil
}

func (s *fakeService) Allow(ctx context.Context, id int64) (model.Party, error) {
	_ = ctx
	if s.err != nil {
		return model.Party{}, s.err
	}
	return model.Party{
		ID:        id,
		IsAllowed: true,
	}, nil
}
