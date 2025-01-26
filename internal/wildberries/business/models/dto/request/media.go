package request

import (
	"encoding/json"
	"gomarketplace_api/internal/wildberries/business/models/dto/response"
)

type MediaRequest struct {
	NmId int      `json:"nmId"`
	Data []string `json:"data"`
}

func NewMediaRequest(nmId int, data []string) *MediaRequest {
	return &MediaRequest{
		NmId: nmId,
		Data: data,
	}
}

func (r *MediaRequest) FromNomenclature(nm response.Nomenclature) *MediaRequest {
	photos := make([]string, len(nm.Photos))

	for i, photo := range nm.Photos {
		photos[i] = photo.Big
	}

	return &MediaRequest{
		NmId: nm.NmID,
		Data: photos,
	}
}
func (r *MediaRequest) ToBytes() ([]byte, error) {
	return json.Marshal(r)
}
