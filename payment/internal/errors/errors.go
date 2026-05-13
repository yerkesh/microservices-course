package errs

import "errors"

var (
	// ErrInvalidUUID возвращается при неверном UUID заказа.
	ErrInvalidUUID = errors.New("неверный формат UUID")
	// ErrInvalidPaymentMethod возвращается при неверном способе оплаты.
	ErrInvalidPaymentMethod = errors.New("неверный метод оплаты")
)
