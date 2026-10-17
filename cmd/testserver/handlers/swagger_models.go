package handlers

import "github.com/Hari-Kiri/virest/utilities"

// DefineRequest is the JSON body for Define.
type DefineRequest struct {
	Option      uint `json:"option" example:"0"`
	StoragePool any  `json:"storagePool"`
}

// UuidOptionRequest is a pool ref (name or UUID) + option body used by several mutating routes.
type UuidOptionRequest struct {
	Uuid   string `json:"uuid" example:"00000000-0000-0000-0000-000000000000"`
	Option uint   `json:"option" example:"0"`
}

// AutostartRequest sets pool autostart.
type AutostartRequest struct {
	Uuid      string `json:"uuid" example:"00000000-0000-0000-0000-000000000000"`
	Autostart bool   `json:"autostart" example:"true"`
}

// FindSourcesRequest discovers pool sources.
type FindSourcesRequest struct {
	Type    utilities.PoolType `json:"type" example:"netfs"`
	SrcSpec any                `json:"srcSpec"`
}

// DefineAsRequest is the body for DefineAs.
type DefineAsRequest struct {
	Option uint                     `json:"option" example:"0"`
	Params utilities.DefineAsParams `json:"params"`
}

// CreateAsRequest is the body for CreateAs.
type CreateAsRequest struct {
	Option uint                     `json:"option" example:"0"`
	Params utilities.CreateAsParams `json:"params"`
}

// FindSourcesAsRequest is the body for FindSourcesAs.
type FindSourcesAsRequest struct {
	Type   utilities.PoolType             `json:"type" example:"iscsi"`
	Params utilities.FindSourcesAsParams `json:"params"`
}

// TokenEnvelope is returned by Authenticate.
type TokenEnvelope struct {
	Response bool `json:"response"`
	Code     int  `json:"code"`
	Data     struct {
		Token string `json:"token"`
	} `json:"data"`
	Error string `json:"error,omitempty"`
}
