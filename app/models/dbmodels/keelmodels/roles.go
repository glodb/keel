package keelmodels

import "time"

type Role struct {
	RoleName   string `json:"roleName"`
	RoleId     int    `json:"roleId"`
	CreatedBy  string `json:"createdBy"`
	IsTemplate bool   `json:"isTemplate"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

func (r *Role) GetQuery() map[string]interface{} {
	return map[string]interface{}{"roleId": r.RoleId}
}

func (ui *Role) GetUpdate() map[string]interface{} {
	return map[string]interface{}{"roleName": ui.RoleName, "roleId": ui.RoleId, "createdBy": ui.CreatedBy, "isTemplate": ui.IsTemplate,
		"createdAt": time.Now().Unix(),
		"updatedAt": time.Now().Unix()}
}

func (ui *Role) MapData(data map[string]interface{}) {

	if val, ok := data["roleName"]; ok {
		ui.RoleName = val.(string)
	}

	if val, ok := data["roleId"]; ok {
		ui.RoleId = val.(int)
	}

	if val, ok := data["createdBy"]; ok {
		ui.CreatedBy = val.(string)
	}

	if val, ok := data["isTemplate"]; ok {
		ui.IsTemplate = val.(bool)
	}

	if val, ok := data["createdAt"]; ok {
		ui.CreatedAt = val.(int64)
	}

	if val, ok := data["updatedAt"]; ok {
		ui.UpdatedAt = val.(int64)
	}
}

func (ui *Role) CleanUp() {
	ui.RoleName = ""
	ui.RoleId = 0
	ui.CreatedBy = ""
	ui.IsTemplate = false
	ui.CreatedAt = 0
	ui.UpdatedAt = 0
}
