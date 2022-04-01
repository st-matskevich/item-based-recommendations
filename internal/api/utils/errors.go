package utils

//response errors readable codes
const (
	AUTHORIZATION_ERROR = "AUTHORIZATION_ERROR"
	SQL_ERROR           = "SQL_ERROR"
	DECODER_ERROR       = "DECODER_ERROR"
	BAD_INPUT           = "BAD_INPUT"
)

//internal errors
const (
	INVALID_INPUT       = "got invalid data"
	INSUFFICIENT_RIGHTS = "user has insufficient rights"
)
