package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hari-Kiri/virest/cmd/testserver/auth"
	"github.com/Hari-Kiri/virest/storagePool"
	"github.com/Hari-Kiri/virest/utilities"
)

// Envelope is the standard JSON response shape for the testserver.
type Envelope struct {
	Response bool        `json:"response"`
	Code     int         `json:"code"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// Deps are shared handler dependencies.
type Deps struct {
	Auth *auth.Store
}

func writeJSON(w http.ResponseWriter, status int, body Envelope) {
	body.Code = status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	msg := "request failed"
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, status, Envelope{Response: false, Error: msg})
}

func (d *Deps) authorize(r *http.Request) (string, error) {
	claims, err := d.Auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		return "", err
	}
	uri := r.Header.Get("Hypervisor-Uri")
	if uri == "" {
		uris := r.Header["Hypervisor-Uri"]
		if len(uris) > 0 {
			uri = uris[0]
		}
	}
	if uri == "" {
		return "", errors.New("hypervisor uri not exist on request header")
	}
	if !d.Auth.AllowHypervisor(claims.Subject, uri) {
		return "", errors.New("hypervisor uri not allowed for user")
	}
	return uri, nil
}

func (d *Deps) connect(r *http.Request) (*storagePool.Connection, error) {
	uri, err := d.authorize(r)
	if err != nil {
		return nil, err
	}
	return storagePool.Connect(uri)
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func query(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func httpStatusFor(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if code, ok := utilities.Code(err); ok {
		switch code {
		case utilities.ErrAuthFailed:
			return http.StatusUnauthorized
		case utilities.ErrNoStoragePool, utilities.ErrNoSupport:
			return http.StatusNotFound
		case utilities.ErrInvalidArg, utilities.ErrInvalidConn:
			return http.StatusBadRequest
		default:
			return http.StatusInternalServerError
		}
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "bearer"), strings.Contains(msg, "credential"), strings.Contains(msg, "token"), strings.Contains(msg, "not allowed"):
		return http.StatusUnauthorized
	case strings.Contains(msg, "hypervisor uri"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

const jsonKeyUUID = "uuid"

func uuidData(uuid string) map[string]string {
	return map[string]string{jsonKeyUUID: uuid}
}

func parseUint(v string) (uint, error) {
	if v == "" {
		return 0, nil
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(n), nil
}
