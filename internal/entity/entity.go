package entity

import "errors"

var (
	ErrNilUUID = errors.New("uuid cannot be nil")
	ErrCheckInFailed = errors.New("already checked in or not found")
)
