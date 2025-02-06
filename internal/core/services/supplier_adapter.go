package services

import "gomarketplace_api/internal/core/models"

// SupplierAdapter определяет, какие операции должен поддерживать адаптер поставщика.
type SupplierAdapter interface {
	// SyncProducts отвечает за синхронизацию списка товаров от поставщика.
	SyncProducts() ([]*models.Product, error)

	// FetchProductDetails возвращает подробную информацию по конкретному товару.
	FetchProductDetails(prodCoreID int) (*models.Product, error)
}
