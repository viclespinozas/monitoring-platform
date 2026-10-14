package registry

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
)

const maxBodyBytes = 1 << 20

// Handler exposes the registry over HTTP.
type Handler struct {
	svc    *Manager
	logger *slog.Logger
}

// NewHandler returns a Handler using svc.
func NewHandler(svc *Manager, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// Routes registers the /api/v1/services routes on mux.
func (h *Handler) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/services", h.create)
	mux.HandleFunc("GET /api/v1/services", h.list)
	mux.HandleFunc("GET /api/v1/services/{id}", h.get)
	mux.HandleFunc("PUT /api/v1/services/{id}", h.update)
	mux.HandleFunc("DELETE /api/v1/services/{id}", h.delete)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	s, err := h.svc.Create(r.Context(), in)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Location", "/api/v1/services/"+s.ID)
	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	s, err := h.svc.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

type listResponse struct {
	Items  []Service `json:"items"`
	Total  int       `json:"total"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{Team: q.Get("team"), Tag: q.Get("tag")}

	var bad []FieldError
	var err error
	if v := q.Get("limit"); v != "" {
		if f.Limit, err = strconv.Atoi(v); err != nil || f.Limit < 1 {
			bad = append(bad, FieldError{"limit", "must be a positive integer"})
		}
	}
	if v := q.Get("offset"); v != "" {
		if f.Offset, err = strconv.Atoi(v); err != nil || f.Offset < 0 {
			bad = append(bad, FieldError{"offset", "must be a non-negative integer"})
		}
	}
	if len(bad) > 0 {
		writeProblem(w, http.StatusUnprocessableEntity, "invalid query parameters", bad)
		return
	}

	f = f.Normalized()
	items, total, err := h.svc.List(r.Context(), f)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, listResponse{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset})
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}
	s, err := h.svc.Update(r.Context(), r.PathValue("id"), in)
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Delete(r.Context(), r.PathValue("id")); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// fail maps domain errors to problem responses. Unknown errors are logged
// and returned as an opaque 500 so internals never leak to clients.
func (h *Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	var ve *ValidationError
	switch {
	case errors.As(err, &ve):
		writeProblem(w, http.StatusUnprocessableEntity, "request validation failed", ve.Fields)
	case errors.Is(err, ErrNotFound):
		writeProblem(w, http.StatusNotFound, "service not found", nil)
	case errors.Is(err, ErrNameTaken):
		writeProblem(w, http.StatusConflict, "a service with this name is already registered", nil)
	default:
		h.logger.ErrorContext(r.Context(), "unexpected error", "method", r.Method, "path", r.URL.Path, "error", err)
		writeProblem(w, http.StatusInternalServerError, "internal server error", nil)
	}
}

func decodeInput(w http.ResponseWriter, r *http.Request) (Input, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var in Input
	if err := dec.Decode(&in); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeProblem(w, http.StatusRequestEntityTooLarge, "request body too large", nil)
			return Input{}, false
		}
		writeProblem(w, http.StatusBadRequest, "malformed JSON body: "+err.Error(), nil)
		return Input{}, false
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		writeProblem(w, http.StatusBadRequest, "body must contain a single JSON object", nil)
		return Input{}, false
	}
	return in, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
