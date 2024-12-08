package storage

import "database/sql"

type NomenclatureRepository struct {
	db *sql.DB
}

func NewNomenclatureRepository(db *sql.DB) *NomenclatureRepository {
	return &NomenclatureRepository{db: db}
}

func (r *NomenclatureRepository) GetSetOfUncreatedItemsWithCategories(accuracy float32, inStock bool, categoryId int) (map[int]interface{}, error) {
	query := `
		SELECT
			whp.global_id, wbc.category, wbc.category_id, wbc.parent_category_id, wbc.parent_category_name
		FROM wholesaler.products AS whp
		LEFT JOIN wildberries.nomenclatures AS wbn ON whp.global_id = wbn.global_id
		LEFT JOIN wildberries.products AS wbp ON whp.global_id = wbp.global_id
		LEFT JOIN wildberries.categories AS wbc ON wbp.category_id = wbc.category_id
		JOIN wholesaler.stocks as whs ON wbp.global_id=whs.global_id
		WHERE wbn.global_id IS NULL
		  AND (wbp.distance < $1 OR wbp.distance IS NULL)
			`

	var rows *sql.Rows
	var err error
	if inStock {
		query += `AND whs.stocks > 0`
	}
	if categoryId > 0 {
		query += `AND wbc.category_id = $2`
		rows, err = r.db.Query(query, accuracy, categoryId)
	} else {
		rows, err = r.db.Query(query, accuracy)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make(map[int]interface{})

	for rows.Next() {
		var globalID int
		var categoryID int
		var categoryName string
		var parentCategoryID int
		var parentCategoryName string
		err = rows.Scan(&globalID, &categoryName, &categoryID, &parentCategoryID, &parentCategoryName)
		if err != nil {
			return nil, err
		}
		items[globalID] = map[interface{}]interface{}{
			"item-id":              globalID,
			"category_id":          categoryID,
			"category":             categoryName,
			"parent_category_id":   parentCategoryID,
			"parent_category_name": parentCategoryName,
		}
	}
	return items, nil
}