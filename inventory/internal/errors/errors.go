package errs

import "errors"

var (
	// ErrPartNotFound возвращается, когда деталь не найдена.
	ErrPartNotFound = errors.New("деталь не найдена")
	// ErrInvalidUUID возвращается при неверном формате UUID.
	ErrInvalidUUID = errors.New("неверный формат UUID")
)
