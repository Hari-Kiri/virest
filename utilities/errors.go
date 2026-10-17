package utilities

import (
	stderrors "errors"
	"fmt"

	"libvirt.org/go/libvirt"
)

// Wrap annotates err with an operation name. Nil input stays nil.
func Wrap(op string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", op, err)
}

// AsLibvirtError extracts a libvirt.Error from err when present.
func AsLibvirtError(err error) (libvirt.Error, bool) {
	var lv libvirt.Error
	if stderrors.As(err, &lv) {
		return lv, true
	}
	return libvirt.Error{}, false
}

// ErrorCode mirrors libvirt's error number type so callers never
// import libvirt.org/go/libvirt.
type ErrorCode = libvirt.ErrorNumber

// Libvirt error code constants mirrored for caller use.
const (
	ErrAuthFailed    ErrorCode = libvirt.ERR_AUTH_FAILED
	ErrNoStoragePool ErrorCode = libvirt.ERR_NO_STORAGE_POOL
	ErrNoSupport     ErrorCode = libvirt.ERR_NO_SUPPORT
	ErrInvalidArg    ErrorCode = libvirt.ERR_INVALID_ARG
	ErrInvalidConn   ErrorCode = libvirt.ERR_INVALID_CONN
)

// Code returns the libvirt error number when err unwraps to libvirt.Error.
func Code(err error) (ErrorCode, bool) {
	lv, ok := AsLibvirtError(err)
	if !ok {
		return 0, false
	}
	return lv.Code, true
}
