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

// ErrorNumber mirrors libvirt's error number type so callers never
// import libvirt.org/go/libvirt.
type CodeError = libvirt.ErrorNumber

// Libvirt error code constants mirrored for caller use.
const (
	ErrAuthFailed    CodeError = libvirt.ERR_AUTH_FAILED
	ErrNoStoragePool CodeError = libvirt.ERR_NO_STORAGE_POOL
	ErrNoSupport     CodeError = libvirt.ERR_NO_SUPPORT
	ErrInvalidArg    CodeError = libvirt.ERR_INVALID_ARG
	ErrInvalidConn   CodeError = libvirt.ERR_INVALID_CONN
)

// Code returns the libvirt error code when err unwraps to libvirt.Error.
func Code(err error) (CodeError, bool) {
	lv, ok := AsLibvirtError(err)
	if !ok {
		return 0, false
	}
	return lv.Code, true
}
