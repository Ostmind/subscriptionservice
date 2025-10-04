package config

import "errors"

var (
	ErrNoServerHost = errors.New("no server host provided")
	ErrNoServerPort = errors.New("no server port provided")
	ErrNoDBHost     = errors.New("no DB host provided")
	ErrNoDBPort     = errors.New("no DB port provided")
	ErrNoDBName     = errors.New("no DB name provided")
	ErrNoDBUser     = errors.New("no DB user provided")
	ErrNoDBPassword = errors.New("no DB password provided")
)
