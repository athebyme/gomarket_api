package models

import (
	"encoding/json"
	"time"
)

// Product представляет товар, агрегируемый из разных источников (поставщиков).
// Эта структура отражает общую информацию о товаре и хранится в таблице core.products.
type Product struct {
	ID         string `json:"id"`
	SupplierID int    `json:"supplier_id"`
	// BaseData содержит базовую информацию о товаре в формате JSON.
	// Здесь могут храниться общие для всех поставщиков поля (например, название, описание, цена и пр.).
	// Для работы с динамическими данными используем json.RawMessage или map[string]interface{}.
	BaseData  json.RawMessage `db:"base_data" json:"base_data"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"` // Дата и время создания записи.
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"` // Дата и время последнего обновления.
}
