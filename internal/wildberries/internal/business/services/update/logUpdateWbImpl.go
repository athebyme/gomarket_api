package update

import "gomarketplace_api/pkg/business/service"

type WbImpl struct {
	*service.DbLog
}

const dbName = "wildberries.changes"

func (wb *WbImpl) UpdateWbChanges() {
	wb.LogUpdate()
}
