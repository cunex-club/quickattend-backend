package entity

import "errors"

var (
	ErrNilUUID = errors.New("uuid cannot be nil")
	ErrAlreadyCommented = errors.New("already commented")
	ErrCheckInTargetNotFound = errors.New("check in target not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
)
