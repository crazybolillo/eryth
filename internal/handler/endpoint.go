package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type Endpoint struct {
	*pgx.Conn
}

type createEndpointRequest struct {
	ID          string   `json:"id"`
	Password    string   `json:"password"`
	Realm       string   `json:"realm,omitempty"`
	Transport   string   `json:"transport,omitempty"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"max_contacts,omitempty"`
	Extension   string   `json:"extension,omitempty"`
	DisplayName string   `json:"display_name"`
}

type listEndpointEntry struct {
	ID          string `json:"id"`
	Extension   string `json:"extension"`
	Context     string `json:"context"`
	DisplayName string `json:"display_name"`
}

type listEndpointsRequest struct {
	Endpoints []listEndpointEntry `json:"endpoints"`
}

func (e *Endpoint) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", e.create)
	r.Get("/list", e.list)
	r.Delete("/{id}", e.delete)

	return r
}

// displayNameFromClid extracts the display name from a Caller ID. It is expected for the Caller ID to be in
// the following format: "Display Name" <username>
// If no display name is found, an empty string is returned.
func displayNameFromClid(callerID string) string {
	if callerID == "" {
		return ""
	}

	start := strings.Index(callerID, `"`)
	if start != 0 {
		return ""
	}

	end := strings.LastIndex(callerID, `"`)
	if end == -1 || end < 1 {
		return ""
	}

	return callerID[1:end]
}

// @Summary List existing endpoints.
// @Param limit query int false "Limit the amount of endpoints returned" default(15)
// @Produce json
// @Success 200 {object} listEndpointsRequest
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoint/list [get]
func (e *Endpoint) list(w http.ResponseWriter, r *http.Request) {
	qlim := r.URL.Query().Get("limit")
	limit := 15
	if qlim != "" {
		conv, err := strconv.Atoi(qlim)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		limit = conv
	}

	queries := sqlc.New(e.Conn)
	rows, err := queries.ListEndpoints(r.Context(), int32(limit))
	if err != nil {
		slog.Error("Query execution failed", slog.String("path", r.URL.Path), slog.String("msg", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rows == nil {
		rows = []sqlc.ListEndpointsRow{}
	}

	endpoints := make([]listEndpointEntry, len(rows))
	for idx := range len(rows) {
		row := rows[idx]
		endpoints[idx] = listEndpointEntry{
			ID:          row.ID,
			Extension:   row.Extension.String,
			Context:     row.Context.String,
			DisplayName: displayNameFromClid(row.Callerid.String),
		}
	}
	response := listEndpointsRequest{
		Endpoints: endpoints,
	}
	content, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}

// @Summary Create a new endpoint.
// @Accept json
// @Param payload body createEndpointRequest true "Endpoint's information"
// @Success 204
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoint [post]
func (e *Endpoint) create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	payload := createEndpointRequest{
		Realm:       "asterisk",
		MaxContacts: 1,
	}

	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := e.Begin(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	queries := sqlc.New(tx)

	hash := md5.Sum([]byte(payload.ID + ":" + payload.Realm + ":" + payload.Password))
	err = queries.NewMD5Auth(r.Context(), sqlc.NewMD5AuthParams{
		ID:       payload.ID,
		Username: db.Text(payload.ID),
		Realm:    db.Text(payload.Realm),
		Md5Cred:  db.Text(hex.EncodeToString(hash[:])),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sid, err := queries.NewEndpoint(r.Context(), sqlc.NewEndpointParams{
		ID:        payload.ID,
		Transport: db.Text(payload.Transport),
		Context:   db.Text(payload.Context),
		Allow:     db.Text(strings.Join(payload.Codecs, ",")),
		Callerid:  db.Text(fmt.Sprintf(`"%s" <%s>`, payload.DisplayName, payload.ID)),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = queries.NewAOR(r.Context(), sqlc.NewAORParams{
		ID:          payload.ID,
		MaxContacts: db.Int4(payload.MaxContacts),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if payload.Extension != "" {
		err = queries.NewExtension(r.Context(), sqlc.NewExtensionParams{
			EndpointID: sid,
			Extension:  db.Text(payload.Extension),
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// @Summary Delete an endpoint and its associated resources.
// @Param id path string true "ID of the endpoint to be deleted"
// @Success 204
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoint/{id} [delete]
func (e *Endpoint) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := e.Begin(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(r.Context())

	queries := sqlc.New(tx)

	err = queries.DeleteEndpoint(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = queries.DeleteAOR(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = queries.DeleteAuth(r.Context(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = tx.Commit(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
