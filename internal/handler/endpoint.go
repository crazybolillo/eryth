package handler

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"net/http"
	"strings"
)

type Endpoint struct {
	*pgx.Conn
}

type createEndpointRequest struct {
	ID          string   `json:"id"`
	Password    string   `json:"password"`
	Realm       string   `json:"realm,omitempty"`
	Transport   string   `json:"transport"`
	Context     string   `json:"context"`
	Codecs      []string `json:"codecs"`
	MaxContacts int32    `json:"max_contacts,omitempty"`
}

func (e *Endpoint) Router() chi.Router {
	r := chi.NewRouter()
	r.Post("/", e.create)

	return r
}

// @Summary Create a new endpoint.
// @Accept json
// @Param payload body createEndpointRequest true "Endpoint's information"
// @Success 204
// @Failure 400
// @Failure 500
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

	err = queries.NewEndpoint(r.Context(), sqlc.NewEndpointParams{
		ID:        payload.ID,
		Transport: db.Text(payload.Transport),
		Context:   db.Text(payload.Context),
		Allow:     db.Text(strings.Join(payload.Codecs, ",")),
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

	err = tx.Commit(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
