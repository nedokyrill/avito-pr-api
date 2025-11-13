package consts

import "time"

const (
	PgxTimeout = 5 * time.Second
	GsTimeout  = 5 * time.Second

	ReadTimeout  = 5 * time.Second
	WriteTimeout = 10 * time.Second
)
