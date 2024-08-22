package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/query"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

const defaultRealm = "asterisk"

type Endpoint struct {
	*pgx.Conn
}

type createEndpointRequest struct {
	ID          string   `json:"id"`
	Password    string   `json:"password"`
	Transport   string   `json:"transport,omitempty"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts,omitempty"`
	Extension   string   `json:"extension,omitempty"`
	DisplayName string   `json:"displayName"`
}

type listEndpointEntry struct {
	Sid         int32  `json:"sid"`
	ID          string `json:"id"`
	Extension   string `json:"extension"`
	Context     string `json:"context"`
	DisplayName string `json:"displayName"`
}

type listEndpointsResponse struct {
	Total     int64               `json:"total"`
	Retrieved int                 `json:"retrieved"`
	Endpoints []listEndpointEntry `json:"endpoints"`
}

type getEndpointResponse struct {
	Sid         int32    `json:"sid"`
	ID          string   `json:"id"`
	DisplayName string   `json:"displayName"`
	Transport   string   `json:"transport"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"maxContacts"`
	Extension   string   `json:"extension"`
}

type updateEndpointRequest struct {
	Password    *string  `json:"password,omitempty"`
	DisplayName *string  `json:"displayName,omitempty"`
	Transport   *string  `json:"transport,omitempty"`
	Context     *string  `json:"context,omitempty"`
	Codecs      []string `json:"codecs,omitempty"`
	MaxContacts *int32   `json:"maxContacts,omitempty"`
	Extension   *string  `json:"extension,omitempty"`
}

func (e *Endpoint) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", e.create)
	r.Get("/", e.list)
	r.Get("/{sid}", e.get)
	r.Delete("/{sid}", e.delete)
	r.Patch("/{sid}", e.update)

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

func hashPassword(user, password, realm string) string {
	hash := md5.Sum([]byte(user + ":" + realm + ":" + password))
	return hex.EncodeToString(hash[:])
}

// @Summary Get information from a specific endpoint.
// @Param sid path int true "Requested endpoint's sid"
// @Produce json
// @Success 200 {object} getEndpointResponse
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [get]
func (e *Endpoint) get(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sid")
	if sid == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id, err := strconv.ParseInt(sid, 10, 32)
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

	row, err := queries.GetEndpointByID(r.Context(), int32(id))
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to retrieve endpoint", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	endpoint := getEndpointResponse{
		Sid:         int32(id),
		ID:          row.ID,
		Transport:   row.Transport.String,
		Context:     row.Context.String,
		Codecs:      strings.Split(row.Allow.String, ","),
		MaxContacts: row.MaxContacts.Int32,
		Extension:   row.Extension.String,
		DisplayName: displayNameFromClid(row.Callerid.String),
	}
	content, err := json.Marshal(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
}

// @Summary List existing endpoints.
// @Param page query int false "Zero based page to fetch" default(0)
// @Param pageSize query int false "Max amount of results to be returned" default(10)
// @Produce json
// @Success 200 {object} listEndpointsResponse
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints [get]
func (e *Endpoint) list(w http.ResponseWriter, r *http.Request) {
	page, err := query.GetIntOr(r.URL.Query(), "page", 0)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pageSize, err := query.GetIntOr(r.URL.Query(), "pageSize", 10)
	if err != nil || page < 0 || pageSize < 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	queries := sqlc.New(e.Conn)
	rows, err := queries.ListEndpoints(r.Context(), sqlc.ListEndpointsParams{
		Limit:  int32(pageSize),
		Offset: int32(page * pageSize),
	})
	if err != nil {
		slog.Error("Query execution failed", slog.String("path", r.URL.Path), slog.String("msg", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rows == nil {
		rows = []sqlc.ListEndpointsRow{}
	}
	total, err := queries.CountEndpoints(r.Context())
	if err != nil {
		slog.Error("Query execution failed", slog.String("path", r.URL.Path), slog.String("msg", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	endpoints := make([]listEndpointEntry, len(rows))
	for idx := range len(rows) {
		row := rows[idx]
		endpoints[idx] = listEndpointEntry{
			Sid:         row.Sid,
			ID:          row.ID,
			Extension:   row.Extension.String,
			Context:     row.Context.String,
			DisplayName: displayNameFromClid(row.Callerid.String),
		}
	}
	response := listEndpointsResponse{
		Total:     total,
		Retrieved: len(rows),
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
// @Success 201 {object} getEndpointResponse
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints [post]
func (e *Endpoint) create(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	payload := createEndpointRequest{
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

	err = queries.NewMD5Auth(r.Context(), sqlc.NewMD5AuthParams{
		ID:       payload.ID,
		Username: db.Text(payload.ID),
		Realm:    db.Text(defaultRealm),
		Md5Cred:  db.Text(hashPassword(payload.ID, payload.Password, defaultRealm)),
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

	// TODO: Duplicate code, same as when fetching endpoint. Probably should put this into a service layer.
	tx, err = e.Begin(r.Context())
	queries = sqlc.New(tx)
	if err != nil {
		slog.Error("Failed to create new transaction", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := queries.GetEndpointByID(r.Context(), sid)
	if err != nil {
		slog.Error(
			"Failed to retrieve created endpoint",
			slog.String("path", r.URL.Path), slog.String("reason", err.Error()), slog.Int("sid", int(sid)),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	endpoint := getEndpointResponse{
		Sid:         sid,
		ID:          res.ID,
		Transport:   res.Transport.String,
		Context:     res.Context.String,
		Codecs:      strings.Split(res.Allow.String, ","),
		MaxContacts: res.MaxContacts.Int32,
		Extension:   res.Extension.String,
		DisplayName: displayNameFromClid(res.Callerid.String),
	}
	content, err := json.Marshal(endpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusCreated)
}

// @Summary Delete an endpoint and its associated resources.
// @Param sid path int true "Sid of the endpoint to be deleted"
// @Success 204
// @Failure 400
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [delete]
func (e *Endpoint) delete(w http.ResponseWriter, r *http.Request) {
	urlSid := chi.URLParam(r, "sid")
	sid, err := strconv.Atoi(urlSid)
	if err != nil || sid <= 0 {
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

	id, err := queries.DeleteEndpoint(r.Context(), int32(sid))
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

// @Summary Update the specified endpoint. Omitted or null fields will remain unchanged.
// @Param sid path int true "Sid of the endpoint to be updated"
// @Success 200 {object} updateEndpointRequest
// @Failure 400
// @Failure 404
// @Failure 500
// @Tags endpoints
// @Router /endpoints/{sid} [patch]
func (e *Endpoint) update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var payload updateEndpointRequest

	err := decoder.Decode(&payload)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlSid := chi.URLParam(r, "sid")
	sid, err := strconv.Atoi(urlSid)
	if err != nil || sid <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tx, err := e.Begin(r.Context())
	if err != nil {
		slog.Error("Failed to start transaction", slog.String("reason", err.Error()), slog.String("path", r.URL.Path))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	queries := sqlc.New(tx)
	endpoint, err := queries.GetEndpointByID(r.Context(), int32(sid))
	if errors.Is(err, pgx.ErrNoRows) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to retrieve endpoint", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	// Sorry for the incoming boilerplate but no dynamic SQL yet
	var patchedEndpoint = sqlc.UpdateEndpointBySidParams{Sid: int32(sid)}
	if payload.DisplayName != nil {
		if *payload.DisplayName == "" {
			patchedEndpoint.Callerid = db.Text("")
		} else {
			patchedEndpoint.Callerid = db.Text(fmt.Sprintf(`"%s" <%s>`, *payload.DisplayName, endpoint.ID))
		}
	} else {
		patchedEndpoint.Callerid = endpoint.Callerid
	}
	if payload.Context != nil {
		patchedEndpoint.Context = db.Text(*payload.Context)
	} else {
		patchedEndpoint.Context = endpoint.Context
	}
	if payload.Transport != nil {
		patchedEndpoint.Transport = db.Text(*payload.Transport)
	} else {
		patchedEndpoint.Transport = endpoint.Transport
	}
	if payload.Codecs != nil {
		patchedEndpoint.Allow = db.Text(strings.Join(payload.Codecs, ","))
	} else {
		patchedEndpoint.Allow = endpoint.Allow
	}
	err = queries.UpdateEndpointBySid(r.Context(), patchedEndpoint)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to update endpoint", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	if payload.MaxContacts != nil {
		err = queries.UpdateAORById(
			r.Context(),
			sqlc.UpdateAORByIdParams{
				ID:          endpoint.ID,
				MaxContacts: db.Int4(*payload.MaxContacts),
			},
		)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to update AOR", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	if payload.Extension != nil {
		err = queries.UpdateExtensionByEndpointId(
			r.Context(),
			sqlc.UpdateExtensionByEndpointIdParams{
				EndpointID: int32(sid),
				Extension:  db.Text(*payload.Extension),
			},
		)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to update extension", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	if payload.Password != nil {
		if len(*payload.Password) < 12 {
			w.WriteHeader(http.StatusBadRequest)
			slog.Info("Invalid password provided", slog.String("path", r.URL.Path))
			return
		}
		err = queries.UpdateMD5AuthById(
			r.Context(),
			sqlc.UpdateMD5AuthByIdParams{
				ID:      endpoint.ID,
				Md5Cred: db.Text(hashPassword(endpoint.ID, *payload.Password, defaultRealm)),
			},
		)
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to update password", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	err = tx.Commit(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to commit update", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		return
	}

	// TODO: Duplicate code, same as when fetching endpoint. Probably should put this into a service layer.
	tx, err = e.Begin(r.Context())
	queries = sqlc.New(tx)
	if err != nil {
		slog.Error("Failed to create new transaction", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := queries.GetEndpointByID(r.Context(), int32(sid))
	if err != nil {
		slog.Error(
			"Failed to retrieve created endpoint",
			slog.String("path", r.URL.Path), slog.String("reason", err.Error()), slog.Int("sid", int(sid)),
		)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := getEndpointResponse{
		Sid:         int32(sid),
		ID:          res.ID,
		Transport:   res.Transport.String,
		Context:     res.Context.String,
		Codecs:      strings.Split(res.Allow.String, ","),
		MaxContacts: res.MaxContacts.Int32,
		Extension:   res.Extension.String,
		DisplayName: displayNameFromClid(res.Callerid.String),
	}
	content, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("Failed to marshall response", slog.String("path", r.URL.Path))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(content)
	if err != nil {
		slog.Error("Failed to write response", slog.String("path", r.URL.Path), slog.String("reason", err.Error()))
	}
	w.WriteHeader(http.StatusOK)
}
