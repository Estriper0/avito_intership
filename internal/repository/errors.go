package repository

import "errors"

var (
	ErrAlreadyExists = errors.New("record exists")
	ErrNotFound      = errors.New("not found")
)
