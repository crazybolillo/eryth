package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/crazybolillo/eryth/internal/model"
	"github.com/crazybolillo/eryth/internal/service"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEndpointAPI(t *testing.T) {
	cases := []struct {
		name string
		test func(handler http.Handler) func(*testing.T)
	}{
		{"Create", MustCreate},
		{"Delete", MustDelete},
		{"Read", MustRead},
		{"Update", MustUpdate},
	}

	conn, err := pgx.Connect(context.Background(), "postgres://go:go@127.0.0.1:54321/eryth")
	if err != nil {
		t.Fatalf(
			"Connection to test database failed: %s. Try running 'make db' and run the tests again",
			err,
		)
	}
	defer func(conn *pgx.Conn, ctx context.Context) {
		err := conn.Close(ctx)
		if err != nil {
			t.Error("Failed to close db connection")
		}
	}(conn, context.Background())

	for _, tt := range cases {
		tx, err := conn.Begin(context.Background())
		if err != nil {
			t.Fatalf("Transaction start failed: %s", err)
		}

		handler := Endpoint{Service: &service.EndpointService{Cursor: tx}}
		t.Run(tt.name, tt.test(handler.Router()))

		err = tx.Rollback(context.Background())
		if err != nil {
			t.Fatalf("Failed to rollback transaction: %s", err)
		}
	}
}

func createEndpoint(t *testing.T, handler http.Handler, endpoint model.NewEndpoint) *httptest.ResponseRecorder {
	payload, err := json.Marshal(endpoint)
	if err != nil {
		t.Errorf("failed to marshal new endpoint: %s", err)
	}

	req := httptest.NewRequest("POST", "/", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

func readEndpoint(handler http.Handler, sid int32) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", fmt.Sprintf("/%d", sid), nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

func updateEndpoint(t *testing.T, handler http.Handler, sid int32, endpoint model.PatchedEndpoint) *httptest.ResponseRecorder {
	payload, err := json.Marshal(endpoint)
	if err != nil {
		t.Errorf("failed to marshal new endpoint: %s", err)
	}
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/%d", sid), bytes.NewReader(payload))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

func parseEndpoint(t *testing.T, content *bytes.Buffer) model.Endpoint {
	var createdEndpoint model.Endpoint
	decoder := json.NewDecoder(content)
	err := decoder.Decode(&createdEndpoint)
	if err != nil {
		t.Errorf("failed to parse endpoint: %s", err)
	}

	return createdEndpoint
}

func MustCreate(handler http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		endpoint := model.NewEndpoint{
			ID:          "zinniaelegans",
			Password:    "verylongandsafepassword",
			Context:     "flowers",
			Codecs:      []string{"ulaw", "g722"},
			Extension:   "1234",
			DisplayName: "Zinnia Elegans",
			MaxContacts: 10,
		}
		res := createEndpoint(t, handler, endpoint)
		if res.Code != http.StatusCreated {
			t.Errorf("invalid http code, got %d, want %d", res.Code, http.StatusCreated)
		}
		got := parseEndpoint(t, res.Body)

		want := model.Endpoint{
			Sid:         got.Sid,
			AccountCode: "zinniaelegans",
			ID:          endpoint.ID,
			DisplayName: endpoint.DisplayName,
			Transport:   endpoint.Transport,
			Context:     endpoint.Context,
			Codecs:      endpoint.Codecs,
			MaxContacts: 10,
			Extension:   "1234",
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Created endpoint does not match request, got %v, want %v", got, want)
		}
	}
}

func MustRead(handler http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		endpoint := model.NewEndpoint{
			ID:          "kiwi",
			Password:    "kiwipassword123",
			Context:     "fruits",
			Codecs:      nil,
			Extension:   "9000",
			DisplayName: "Blue Kiwi",
		}
		res := createEndpoint(t, handler, endpoint)
		want := parseEndpoint(t, res.Body)

		res = readEndpoint(handler, want.Sid)
		if res.Code != http.StatusOK {
			t.Errorf("invalid http code, got %d, want %d", res.Code, http.StatusOK)
		}
		got := parseEndpoint(t, res.Body)

		if !reflect.DeepEqual(want, got) {
			t.Errorf("read endpoint does not match want, got %v, want %v", got, want)
		}
	}
}

func MustDelete(handler http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		endpoint := model.NewEndpoint{
			ID:          "testuser",
			Password:    "testpassword123$",
			Context:     "internal",
			Codecs:      nil,
			Extension:   "4000",
			DisplayName: "Mr. Test User",
		}

		res := createEndpoint(t, handler, endpoint)
		createdEndpoint := parseEndpoint(t, res.Body)

		req := httptest.NewRequest("DELETE", fmt.Sprintf("/%d", createdEndpoint.Sid), nil)
		res = httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		if res.Code != http.StatusNoContent {
			t.Errorf("invalid http code, got %d, want %d", res.Code, http.StatusNoContent)
		}

		res = readEndpoint(handler, createdEndpoint.Sid)
		if res.Code != http.StatusNotFound {
			t.Errorf("invalid http code, got %d, want %d", res.Code, http.StatusNotFound)
		}
	}
}

func MustUpdate(handler http.Handler) func(*testing.T) {
	return func(t *testing.T) {
		endpoint := model.NewEndpoint{
			ID:          "big_chungus",
			Password:    "big_chungus_password",
			Context:     "memes",
			Codecs:      []string{"ulaw", "opus"},
			Extension:   "5061",
			DisplayName: "Big Chungus",
		}
		res := createEndpoint(t, handler, endpoint)
		want := parseEndpoint(t, res.Body)
		want.MaxContacts = 5
		want.Extension = "6072"

		res = updateEndpoint(t, handler, want.Sid, model.PatchedEndpoint{
			MaxContacts: &want.MaxContacts,
			Extension:   &want.Extension,
		})
		if res.Code != http.StatusOK {
			t.Errorf("invalid http code, got %d, want %d", res.Code, http.StatusOK)
		}
		got := parseEndpoint(t, res.Body)

		if !reflect.DeepEqual(got, want) {
			t.Errorf("inconsistent update result, got %v, want %v", got, want)
		}
	}
}
