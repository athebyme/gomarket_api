package get

import "time"

type DBUpdateTime struct {
	Id         int       `json:"update_id"`
	GlobalID   int       `json:"global_id"`
	NmID       int       `json:"nm_id"`
	ImtID      int       `json:"imt_id"`
	NmUUID     string    `json:"nm_uuid"`
	VendorCode string    `json:"vendor_code"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
