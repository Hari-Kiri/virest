package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Hari-Kiri/virest/storagePool"
	"github.com/Hari-Kiri/virest/utilities"
)

// Authenticate issues a JWT from HTTP Basic credentials.
//
//	@Summary		Issue JWT
//	@Description	Authenticate with HTTP Basic and receive a Bearer JWT
//	@Tags			auth
//	@Produce		json
//	@Security		BasicAuth
//	@Success		200	{object}	TokenEnvelope
//	@Failure		401	{object}	Envelope
//	@Router			/storage-pool/authenticate [get]
func (d *Deps) Authenticate(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		writeErr(w, http.StatusUnauthorized, errBasicAuthRequired)
		return
	}
	token, err := d.Auth.AuthenticateBasic(username, password)
	if err != nil {
		writeErr(w, http.StatusUnauthorized, err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{
		Response: true,
		Data:     map[string]string{"token": token},
	})
}

var errBasicAuthRequired = handlerError("basic auth required")

type handlerError string

func (e handlerError) Error() string { return string(e) }

// List lists storage pools.
//
//	@Summary		List storage pools
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI (must be allowlisted)"
//	@Param			option			query		int		false	"List flags"
//	@Param			inactive		query		int		false	"XML flags for inactive pools"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/list [get]
func (d *Deps) List(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()

	option, err := parseUint(query(r, "option"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	inactive, err := parseUint(query(r, "inactive"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	result, err := conn.List(utilities.ListFlags(option), utilities.XMLFlags(inactive))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Info returns volatile pool info.
//
//	@Summary		Get pool info
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	true	"Pool name or UUID"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/info [get]
func (d *Deps) Info(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	result, err := conn.Info(query(r, "uuid"))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Detail returns full pool dumpxml model (HTTP alias of Dump).
//
//	@Summary		Get pool detail
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	true	"Pool name or UUID"
//	@Param			option			query		int		false	"XML flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/detail [get]
func (d *Deps) Detail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	option, err := parseUint(query(r, "option"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	result, err := conn.Dump(query(r, "uuid"), utilities.XMLFlags(option))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Capabilities returns pool capabilities.
//
//	@Summary		Get storage pool capabilities
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/capabilities [get]
func (d *Deps) Capabilities(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	result, err := conn.Capabilities()
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Define defines a new storage pool.
//
//	@Summary		Define a storage pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string			true	"Libvirt URI"
//	@Param			body			body		DefineRequest	true	"Pool definition"
//	@Success		201				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/define [post]
func (d *Deps) Define(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body struct {
		Option      utilities.DefineFlags `json:"option"`
		StoragePool utilities.StoragePool          `json:"storagePool"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	uuid, err := conn.Define(body.StoragePool, body.Option)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, Envelope{Response: true, Data: uuidData(uuid)})
}

type uuidOptionRequest = UuidOptionRequest

type autostartRequest = AutostartRequest

func (d *Deps) withUUIDConn(w http.ResponseWriter, r *http.Request, method string, fn func(*storagePool.Connection, string) error) {
	if r.Method != method {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body uuidOptionRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := fn(conn, body.Uuid); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(body.Uuid)})
}

// Build builds pool storage.
//
//	@Summary		Build pool storage
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool name or UUID and build flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/build [patch]
func (d *Deps) Build(w http.ResponseWriter, r *http.Request) {
	var body uuidOptionRequest
	if r.Method != http.MethodPatch {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := conn.Build(body.Uuid, utilities.BuildFlags(body.Option)); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(body.Uuid)})
}

// Create starts a pool.
//
//	@Summary		Create (start) a pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool name or UUID and create flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/create [patch]
func (d *Deps) Create(w http.ResponseWriter, r *http.Request) {
	// HTTP alias of Start (library Connection.Create removed).
	d.Start(w, r)
}

// Destroy stops a pool.
//
//	@Summary		Destroy (stop) a pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool UUID"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/destroy [patch]
func (d *Deps) Destroy(w http.ResponseWriter, r *http.Request) {
	d.withUUIDConn(w, r, http.MethodPatch, func(c *storagePool.Connection, uuid string) error {
		return c.Destroy(uuid)
	})
}

// Undefine removes a pool definition.
//
//	@Summary		Undefine a pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool UUID"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/undefine [delete]
func (d *Deps) Undefine(w http.ResponseWriter, r *http.Request) {
	d.withUUIDConn(w, r, http.MethodDelete, func(c *storagePool.Connection, uuid string) error {
		return c.Undefine(uuid)
	})
}

// Delete deletes pool resources.
//
//	@Summary		Delete pool resources
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool UUID and delete flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/delete [delete]
func (d *Deps) Delete(w http.ResponseWriter, r *http.Request) {
	var body uuidOptionRequest
	if r.Method != http.MethodDelete {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := conn.Delete(body.Uuid, utilities.DeleteFlags(body.Option)); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true})
}

// Refresh refreshes pool volumes.
//
//	@Summary		Refresh pool volumes
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	false	"Pool UUID (or JSON body)"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/refresh [post]
func (d *Deps) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	uuid := query(r, "uuid")
	if uuid == "" {
		var body struct {
			Uuid string `json:"uuid"`
		}
		if err := decodeJSON(r, &body); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		uuid = body.Uuid
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := conn.Refresh(uuid); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(uuid)})
}

// Autostart sets pool autostart.
//
//	@Summary		Set pool autostart
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		AutostartRequest	true	"Autostart settings"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/autostart [patch]
func (d *Deps) Autostart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body autostartRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := conn.Autostart(body.Uuid, body.Autostart); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(body.Uuid)})
}

// FindSources discovers pool sources.
//
//	@Summary		Find storage pool sources
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		FindSourcesRequest	true	"Pool type and source spec"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/find-storage-pool-sources [post]
func (d *Deps) FindSources(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body struct {
		Type    utilities.PoolType     `json:"type"`
		SrcSpec storagePool.SourceSpec `json:"srcSpec"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	result, err := conn.FindSources(body.Type, body.SrcSpec)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Event waits for one event or streams SSE when stream=1.
//
//	@Summary		Wait for pool event (or SSE stream)
//	@Tags			events
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	true	"Pool UUID"
//	@Param			types			query		int		false	"0=lifecycle, 1=refresh"
//	@Param			stream			query		string	false	"Set to 1 for SSE"
//	@Param			timeout			query		int		false	"SSE timeout seconds"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/event [get]
func (d *Deps) Event(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()

	uuid := query(r, "uuid")
	kindU, err := parseUint(query(r, "types"))
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	kind := storagePool.EventKind(kindU)
	stream := query(r, "stream") == "1"

	if stream {
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeErr(w, http.StatusInternalServerError, handlerError("streaming unsupported"))
			return
		}
		timeoutSec, _ := strconv.Atoi(query(r, "timeout"))
		ctx := r.Context()
		if timeoutSec > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSec)*time.Second)
			defer cancel()
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)
		flusher.Flush()
		err := conn.StreamEvents(ctx, uuid, kind, func(ev storagePool.Event) error {
			payload, _ := json.Marshal(ev)
			if _, err := w.Write([]byte("data: " + string(payload) + "\n\n")); err != nil {
				return err
			}
			flusher.Flush()
			return nil
		})
		if err != nil && ctx.Err() == nil {
			tembo := []byte("event: error\ndata: " + err.Error() + "\n\n")
			_, _ = w.Write(tembo)
			flusher.Flush()
		}
		return
	}

	result, err := conn.WaitEvent(uuid, kind)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// CreateTransient creates a transient pool from a model.
//
//	@Summary		Create transient pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string			true	"Libvirt URI"
//	@Param			body			body		DefineRequest	true	"Pool model and flags"
//	@Success		201				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/create-transient [post]
func (d *Deps) CreateTransient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body struct {
		Option      utilities.CreateFlags `json:"option"`
		StoragePool utilities.StoragePool          `json:"storagePool"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	uuid, err := conn.CreateTransient(body.StoragePool, body.Option)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, Envelope{Response: true, Data: uuidData(uuid)})
}

// DefineAs defines a pool from virsh-style -as params.
//
//	@Summary		Define pool from args
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string			true	"Libvirt URI"
//	@Param			body			body		DefineAsRequest	true	"Define-as params"
//	@Success		201				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/define-as [post]
func (d *Deps) DefineAs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body DefineAsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	uuid, err := conn.DefineAs(body.Params, utilities.DefineFlags(body.Option))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, Envelope{Response: true, Data: uuidData(uuid)})
}

// CreateAs creates a transient pool from virsh-style -as params.
//
//	@Summary		Create transient pool from args
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string			true	"Libvirt URI"
//	@Param			body			body		CreateAsRequest	true	"Create-as params"
//	@Success		201				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/create-as [post]
func (d *Deps) CreateAs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body CreateAsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	uuid, err := conn.CreateAs(body.Params, utilities.CreateFlags(body.Option))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusCreated, Envelope{Response: true, Data: uuidData(uuid)})
}

// FindSourcesAs discovers sources from typed -as params.
//
//	@Summary		Find sources from args
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string					true	"Libvirt URI"
//	@Param			body			body		FindSourcesAsRequest	true	"Find-sources-as params"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/find-sources-as [post]
func (d *Deps) FindSourcesAs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	var body FindSourcesAsRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	result, err := conn.FindSourcesAs(body.Type, body.Params)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: result})
}

// Start starts a pool (virsh pool-start).
//
//	@Summary		Start a pool
//	@Tags			pools
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string				true	"Libvirt URI"
//	@Param			body			body		UuidOptionRequest	true	"Pool name or UUID and flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/start [patch]
func (d *Deps) Start(w http.ResponseWriter, r *http.Request) {
	var body uuidOptionRequest
	if r.Method != http.MethodPatch {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	if err := conn.Start(body.Uuid, utilities.CreateFlags(body.Option)); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(body.Uuid)})
}

// Dump returns pool dumpxml model (virsh pool-dumpxml).
//
//	@Summary		Dump pool description
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	true	"Pool name or UUID"
//	@Param			option			query		int		false	"XML flags"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/dump [get]
func (d *Deps) Dump(w http.ResponseWriter, r *http.Request) {
	d.Detail(w, r)
}

// Name returns the pool name for a ref.
//
//	@Summary		Pool name from ref
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			uuid			query		string	true	"Pool name or UUID"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/name [get]
func (d *Deps) Name(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	name, err := conn.Name(query(r, "uuid"))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: map[string]string{"name": name}})
}

// UUIDByName returns the pool UUID for a name.
//
//	@Summary		Pool UUID from name
//	@Tags			pools
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI"
//	@Param			name			query		string	true	"Pool name"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/uuid-by-name [get]
func (d *Deps) UUIDByName(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, handlerError("method not allowed"))
		return
	}
	conn, err := d.connect(r)
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	defer conn.Close()
	uuid, err := conn.UUIDByName(query(r, "name"))
	if err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: uuidData(uuid)})
}

// ProcessUID returns the testserver process UID (not a hypervisor API).
//
//	@Summary		Process UID of the testserver
//	@Tags			meta
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI (allowlist check)"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/get-uid [get]
func (d *Deps) ProcessUID(w http.ResponseWriter, r *http.Request) {
	if _, err := d.authorize(r); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: map[string]int{"uid": processUID()}})
}

// ProcessGID returns the testserver process GID (not a hypervisor API).
//
//	@Summary		Process GID of the testserver
//	@Tags			meta
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Hypervisor-Uri	header		string	true	"Libvirt URI (allowlist check)"
//	@Success		200				{object}	Envelope
//	@Failure		401				{object}	Envelope
//	@Router			/storage-pool/get-gid [get]
func (d *Deps) ProcessGID(w http.ResponseWriter, r *http.Request) {
	if _, err := d.authorize(r); err != nil {
		writeErr(w, httpStatusFor(err), err)
		return
	}
	writeJSON(w, http.StatusOK, Envelope{Response: true, Data: map[string]int{"gid": processGID()}})
}
