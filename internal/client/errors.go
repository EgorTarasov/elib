package client

import "errors"

var (
	ErrCredentialsConfgNotExist = errors.New("credential config not exist")
	ErrCannotLoadCredentials    = errors.New("can't load credentials")
	ErrAuthFailed               = errors.New("can't auth with saved credentials")
)
