package update

import (
	"fmt"
	"gomarketplace_api/internal/wildberries/internal/business/models/dto/response"
	"gomarketplace_api/internal/wildberries/internal/business/models/get"
)

func UpdateProduct(nmID int, imtID int, productData map[string]interface{}) error {
	// 1. Получить текущие характеристики товара с Wildberries
	existingProduct, err := GetProductFromWB(nmID) // Функция для получения данных о товаре
	if err != nil {
		return err
	}

	// 2. Получить характеристики для категории товара
	categoryCharcs, err := GetCategoryCharacteristics(existingProduct.SubjectID)
	if err != nil {
		return err
	}

	newCharacteristics := make([]response.Charc, 0)
	for _, charc := range categoryCharcs {
		if val, ok := productData[charc.Name]; ok {
			newCharacteristics = append(newCharacteristics, response.Charc{
				Id:    charc.CharcID,
				Name:  charc.Name,
				Value: []string{fmt.Sprintf("%v", val)},
			})
			delete(productData, charc.Name)
		} else if charc.Required {
			for _, existingCharc := range existingProduct.Characteristics {
				if existingCharc.Id == charc.CharcID {
					newCharacteristics = append(newCharacteristics, existingCharc)
					break
				}
			}
		}
	}

	// 4. Добавить новые характеристики, которых нет в categoryCharcs (если нужно)
	for key, value := range productData {
		newCharacteristics = append(newCharacteristics, response.Charc{
			Name:  key,
			Value: []string{fmt.Sprintf("%v", value)},
		})
	}

	// ...  обновление других полей товара ...

	// 5. Отправить обновленные данные на Wildberries
	err = SendUpdatedProductToWB(nmID, imtID, newCharacteristics)
	if err != nil {
		return err
	}

	return nil
}

func GetCategoryCharacteristics(subjectID int) ([]get.FullCharcsInfo, error) {
	panic("Implement me")
}

func SendUpdatedProductToWB(nmID int, imtID int, characteristics []response.Charc) error {
	panic("Implement me")
}

func GetProductFromWB(nmID int) (*response.Nomenclature, error) {
	panic("Implement me")
}
