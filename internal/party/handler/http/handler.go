package http

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strconv"

	"Parties/internal/party/model"
	"Parties/internal/party/repository"
	partyservice "Parties/internal/party/service"

	"github.com/go-chi/chi/v5"
)

type Service interface {
	Create(ctx context.Context, input model.CreatePartyInput) (model.Party, error)
	GetByID(ctx context.Context, id int64) (model.Party, error)
	List(ctx context.Context, filter model.ListPartiesFilter) ([]model.Party, error)
	Update(ctx context.Context, input model.UpdatePartyInput) (model.Party, error)
	Delete(ctx context.Context, id int64) error
	Ban(ctx context.Context, id int64, reason string) (model.Party, error)
	Allow(ctx context.Context, id int64) (model.Party, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/parties", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
		r.Patch("/{id}/ban", h.Ban)
		r.Patch("/{id}/allow", h.Allow)
	})
}

func (h *Handler) Create(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request createPartyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid JSON body")
		return
	}

	party, err := h.service.Create(r.Context(), model.CreatePartyInput{
		Name:        request.Name,
		Description: request.Description,
		Promises:    request.Promises,
		Image:       request.Image,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusCreated, toPartyResponse(party))
}

func (h *Handler) GetByID(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	party, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, toPartyResponse(party))
}

func (h *Handler) List(w nethttp.ResponseWriter, r *nethttp.Request) {
	limit, err := intQuery(r, "limit", 50)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid limit")
		return
	}

	offset, err := intQuery(r, "offset", 0)
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid offset")
		return
	}

	parties, err := h.service.List(r.Context(), model.ListPartiesFilter{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := make([]partyResponse, 0, len(parties))
	for _, party := range parties {
		response = append(response, toPartyResponse(party))
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func (h *Handler) Update(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var request updatePartyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid JSON body")
		return
	}

	party, err := h.service.Update(r.Context(), model.UpdatePartyInput{
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		Promises:    request.Promises,
		Image:       request.Image,
		IsAllowed:   request.IsAllowed,
		BanReason:   request.BanReason,
	})
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, toPartyResponse(party))
}

func (h *Handler) Delete(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		writeServiceError(w, err)
		return
	}

	w.WriteHeader(nethttp.StatusNoContent)
}

func (h *Handler) Ban(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	var request banPartyRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, nethttp.StatusBadRequest, "invalid JSON body")
		return
	}

	party, err := h.service.Ban(r.Context(), id, request.Reason)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, toPartyResponse(party))
}

func (h *Handler) Allow(w nethttp.ResponseWriter, r *nethttp.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	party, err := h.service.Allow(r.Context(), id)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, nethttp.StatusOK, toPartyResponse(party))
}

type createPartyRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Promises    string `json:"promises"`
	Image       string `json:"image"`
}

type updatePartyRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Promises    string  `json:"promises"`
	Image       string  `json:"image"`
	IsAllowed   bool    `json:"is_allowed"`
	BanReason   *string `json:"ban_reason"`
}

type banPartyRequest struct {
	Reason string `json:"reason"`
}

type partyResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Promises    string  `json:"promises"`
	Image       string  `json:"image"`
	IsAllowed   bool    `json:"is_allowed"`
	BanReason   *string `json:"ban_reason,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toPartyResponse(party model.Party) partyResponse {
	return partyResponse{
		ID:          party.ID,
		Name:        party.Name,
		Description: party.Description,
		Promises:    party.Promises,
		Image:       party.Image,
		IsAllowed:   party.IsAllowed,
		BanReason:   party.BanReason,
	}
}

func parseID(w nethttp.ResponseWriter, r *nethttp.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, nethttp.StatusBadRequest, "invalid party id")
		return 0, false
	}

	return id, true
}

func intQuery(r *nethttp.Request, name string, defaultValue int) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue, nil
	}

	return strconv.Atoi(value)
}

func decodeJSON(r *nethttp.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeServiceError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, partyservice.ErrInvalidInput):
		writeError(w, nethttp.StatusBadRequest, "invalid party input")
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, nethttp.StatusNotFound, "party not found")
	default:
		writeError(w, nethttp.StatusInternalServerError, "internal server error")
	}
}

func writeError(w nethttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func writeJSON(w nethttp.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
