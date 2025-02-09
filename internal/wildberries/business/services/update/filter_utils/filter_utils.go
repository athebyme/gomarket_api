package filter_utils

import "fmt"

// FilterData обобщённая функция фильтрации с двумя типовыми параметрами.
// U — тип данных, который возвращается функцией fetchFunc.
// T — тип данных, который возвращается функцией transformFunc.
func FilterData[T any, U any](
	ids []int,
	fetchFunc func(ids []int) (map[int]U, error),
	transformFunc func(id int, data U) (T, bool, error),
) (map[int]T, error) {
	idSet := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	dataMap, err := fetchFunc(ids)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	if len(ids) == 0 {
		result := make(map[int]T, len(dataMap))
		for id, data := range dataMap {
			transformed, _, err := transformFunc(id, data)
			if err != nil {
				return nil, fmt.Errorf("failed to transform data for ID %d: %w", id, err)
			}
			result[id] = transformed
		}
		return result, nil
	}

	filtered := make(map[int]T, len(ids))
	for id, data := range dataMap {
		if _, exists := idSet[id]; exists {
			transformed, include, err := transformFunc(id, data)
			if err != nil {
				return nil, fmt.Errorf("failed to transform data for ID %d: %w", id, err)
			}
			if include {
				filtered[id] = transformed
			}
		}
	}

	return filtered, nil
}
