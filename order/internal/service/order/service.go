package order

// Service реализует бизнес-логику заказов.
type Service struct {
	txManager       TxManager
	orderRepo       OrderRepository
	inventoryClient InventoryClient
	paymentClient   PaymentClient
}

// New создаёт сервис заказов.
func New(
	txManager TxManager,
	orderRepo OrderRepository,
	inventoryClient InventoryClient,
	paymentClient PaymentClient,
) *Service {
	return &Service{
		txManager:       txManager,
		orderRepo:       orderRepo,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
	}
}
