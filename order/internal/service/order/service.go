package order

// Service реализует бизнес-логику заказов.
type Service struct {
	orderRepo       OrderRepository
	inventoryClient InventoryClient
	paymentClient   PaymentClient
}

// New создаёт сервис заказов.
func New(orderRepo OrderRepository, inventoryClient InventoryClient, paymentClient PaymentClient) *Service {
	return &Service{
		orderRepo:       orderRepo,
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
	}
}
