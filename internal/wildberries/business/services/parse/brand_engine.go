package parse

type BrandService interface {
	IsBanned(brand string) bool
}

type BrandServiceWildberries struct {
	banned map[string]struct{}
}

func NewBrandServiceWildberries(banned []string) *BrandServiceWildberries {
	bannedMap := make(map[string]struct{}, len(banned))
	for _, b := range banned {
		bannedMap[b] = struct{}{}
	}
	return &BrandServiceWildberries{banned: bannedMap}
}

func (s *BrandServiceWildberries) IsBanned(brand string) bool {
	_, ok := s.banned[brand]
	return ok
}
