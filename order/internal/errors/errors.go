package errs

import "errors"

var (
	// ErrOrderNotFound возвращается, когда заказ не найден.
	ErrOrderNotFound = errors.New("заказ не найден")
	// ErrOrderAlreadyPaid возвращается при повторной оплате или отмене оплаченного заказа.
	ErrOrderAlreadyPaid = errors.New("заказ уже оплачен")
	// ErrOrderCancelled возвращается при действии над отменённым заказом.
	ErrOrderCancelled = errors.New("заказ отменён")
	// ErrPaymentInProgress возвращается при конкурентной операции над заказом в оплате.
	ErrPaymentInProgress = errors.New("оплата заказа уже обрабатывается")
	// ErrPartNotFound возвращается, когда деталь не найдена.
	ErrPartNotFound = errors.New("деталь не найдена")
	// ErrOutOfStock возвращается, когда деталь отсутствует на складе.
	ErrOutOfStock = errors.New("деталь отсутствует на складе")
	// ErrInvalidUUID возвращается при неверном формате UUID.
	ErrInvalidUUID = errors.New("неверный формат UUID")
	// ErrInvalidPaymentMethod возвращается при неверном способе оплаты.
	ErrInvalidPaymentMethod = errors.New("неверный метод оплаты")
)
