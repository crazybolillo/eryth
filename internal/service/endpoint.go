package service

import (
	"cmp"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/crazybolillo/eryth/internal/db"
	"github.com/crazybolillo/eryth/internal/sqlc"
	"github.com/crazybolillo/eryth/pkg/model"
	"strings"
)

const defaultRealm = "asterisk"

type EndpointService struct {
	Cursor
}

func hashPassword(user, password, realm string) string {
	hash := md5.Sum([]byte(user + ":" + realm + ":" + password))
	return hex.EncodeToString(hash[:])
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

func parseMediaEncryption(value string) (sqlc.NullPjsipMediaEncryptionValues, error) {
	switch value {
	case "sdes":
		fallthrough
	case "dtls":
		return sqlc.NullPjsipMediaEncryptionValues{
			PjsipMediaEncryptionValues: sqlc.PjsipMediaEncryptionValues(value),
			Valid:                      true,
		}, nil
	case "":
		return sqlc.NullPjsipMediaEncryptionValues{}, nil
	default:
		return sqlc.NullPjsipMediaEncryptionValues{}, fmt.Errorf("invalid value for media_encryption '%s'", value)
	}
}

func (e *EndpointService) Create(ctx context.Context, payload model.NewEndpoint) (model.Endpoint, error) {
	tx, err := e.Begin(ctx)
	if err != nil {
		return model.Endpoint{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlc.New(tx)
	err = queries.NewMD5Auth(ctx, sqlc.NewMD5AuthParams{
		ID:       payload.ID,
		Username: db.Text(payload.ID),
		Realm:    db.Text(defaultRealm),
		Md5Cred:  db.Text(hashPassword(payload.ID, payload.Password, defaultRealm)),
	})
	if err != nil {
		return model.Endpoint{}, err
	}

	natValue := sqlc.NullAstBoolValues{AstBoolValues: "no", Valid: true}
	if payload.Nat {
		natValue.AstBoolValues = "yes"
	}

	mediaValue, err := parseMediaEncryption(payload.Encryption)
	if err != nil {
		return model.Endpoint{}, err
	}

	sid, err := queries.NewEndpoint(ctx, sqlc.NewEndpointParams{
		ID:              payload.ID,
		Accountcode:     db.Text(cmp.Or(payload.AccountCode, payload.ID)),
		Transport:       db.Text(payload.Transport),
		Context:         db.Text(payload.Context),
		Allow:           db.Text(strings.Join(payload.Codecs, ",")),
		Callerid:        db.Text(fmt.Sprintf(`"%s" <%s>`, payload.DisplayName, payload.ID)),
		Nat:             natValue,
		MediaEncryption: mediaValue,
	})
	if err != nil {
		return model.Endpoint{}, err
	}

	err = queries.NewAOR(ctx, sqlc.NewAORParams{
		ID:          payload.ID,
		MaxContacts: db.Int4(payload.MaxContacts),
	})
	if err != nil {
		return model.Endpoint{}, err
	}

	if payload.Extension != "" {
		err = queries.NewExtension(ctx, sqlc.NewExtensionParams{
			EndpointID: sid,
			Extension:  db.Text(payload.Extension),
		})
		if err != nil {
			return model.Endpoint{}, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Endpoint{}, err
	}

	return e.Read(ctx, sid)
}

func (e *EndpointService) Read(ctx context.Context, sid int32) (model.Endpoint, error) {
	queries := sqlc.New(e.Cursor)
	row, err := queries.GetEndpointByID(ctx, sid)
	if err != nil {
		return model.Endpoint{}, err
	}

	endpoint := model.Endpoint{
		Sid:         sid,
		ID:          row.ID,
		AccountCode: row.Accountcode.String,
		DisplayName: displayNameFromClid(row.Callerid.String),
		Transport:   row.Transport.String,
		Context:     row.Context.String,
		Codecs:      strings.Split(row.Allow.String, ","),
		MaxContacts: row.MaxContacts.Int32,
		Extension:   row.Extension.String,
		Nat:         row.Nat.Bool,
		Encryption:  string(row.MediaEncryption.PjsipMediaEncryptionValues),
	}
	return endpoint, nil
}

func (e *EndpointService) Update(ctx context.Context, sid int32, payload model.PatchedEndpoint) (model.Endpoint, error) {
	tx, err := e.Begin(ctx)
	if err != nil {
		return model.Endpoint{}, err
	}
	defer tx.Rollback(ctx)

	queries := sqlc.New(tx)
	endpoint, err := queries.GetEndpointByID(ctx, sid)
	if err != nil {
		return model.Endpoint{}, err
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
	if payload.Nat != nil {
		patchedEndpoint.Nat = sqlc.NullAstBoolValues{AstBoolValues: "yes", Valid: *payload.Nat}
	} else {
		patchedEndpoint.Nat = sqlc.NullAstBoolValues{AstBoolValues: "yes", Valid: endpoint.Nat.Bool}
	}
	if payload.Encryption != nil {
		patchedEndpoint.MediaEncryption, err = parseMediaEncryption(*payload.Encryption)
		if err != nil {
			return model.Endpoint{}, err
		}
	} else {
		patchedEndpoint.MediaEncryption = endpoint.MediaEncryption
	}

	err = queries.UpdateEndpointBySid(ctx, patchedEndpoint)
	if err != nil {
		return model.Endpoint{}, err
	}

	if payload.MaxContacts != nil {
		err = queries.UpdateAORById(
			ctx,
			sqlc.UpdateAORByIdParams{
				ID:          endpoint.ID,
				MaxContacts: db.Int4(*payload.MaxContacts),
			},
		)
	}
	if err != nil {
		return model.Endpoint{}, err
	}

	if payload.Extension != nil {
		err = queries.UpdateExtensionByEndpointId(
			ctx,
			sqlc.UpdateExtensionByEndpointIdParams{
				EndpointID: sid,
				Extension:  db.Text(*payload.Extension),
			},
		)
	}
	if err != nil {
		return model.Endpoint{}, err
	}

	if payload.Password != nil {
		err = queries.UpdateMD5AuthById(
			ctx,
			sqlc.UpdateMD5AuthByIdParams{
				ID:      endpoint.ID,
				Md5Cred: db.Text(hashPassword(endpoint.ID, *payload.Password, defaultRealm)),
			},
		)
	}
	if err != nil {
		return model.Endpoint{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return model.Endpoint{}, err
	}

	return e.Read(ctx, sid)
}

func (e *EndpointService) Delete(ctx context.Context, sid int32) error {
	tx, err := e.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlc.New(tx)

	id, err := queries.DeleteEndpoint(ctx, sid)
	if err != nil {
		return err
	}

	err = queries.DeleteAOR(ctx, id)
	if err != nil {
		return err
	}

	err = queries.DeleteAuth(ctx, id)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (e *EndpointService) Paginate(ctx context.Context, page, size int) (model.EndpointPage, error) {
	queries := sqlc.New(e.Cursor)
	rows, err := queries.ListEndpoints(ctx, sqlc.ListEndpointsParams{
		Limit:  int32(size),
		Offset: int32(page * size),
	})
	if err != nil {
		return model.EndpointPage{}, err
	}

	count, err := queries.CountEndpoints(ctx)
	if err != nil {
		return model.EndpointPage{}, err
	}

	if rows == nil {
		rows = []sqlc.ListEndpointsRow{}
	}

	endpoints := make([]model.EndpointPageEntry, len(rows))
	for idx := range len(rows) {
		row := rows[idx]
		endpoints[idx] = model.EndpointPageEntry{
			Sid:         row.Sid,
			ID:          row.ID,
			Extension:   row.Extension.String,
			Context:     row.Context.String,
			DisplayName: displayNameFromClid(row.Callerid.String),
		}
	}

	return model.EndpointPage{
		Total:     count,
		Retrieved: len(rows),
		Endpoints: endpoints,
	}, nil
}
