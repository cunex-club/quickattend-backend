package entity

import "errors"

var (
	ErrNilUUID = errors.New("uuid cannot be nil")
	ErrAlreadyCheckedIn = errors.New("already checked in")
	ErrCheckInTargetNotFound = errors.New("check in target not found")
)
