package keelmodels

type RBAC struct {
	ApiName   string `json:"apiName"`
	Role      string `json:"role"`
	CreatedBy string `json:"createdBy"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

func (r *RBAC) GetQuery() map[string]interface{} {
	return map[string]interface{}{"role": r.Role, "apiName": r.ApiName}
}

func (r *RBAC) GetUpdate() map[string]interface{} {
	return map[string]interface{}{"role": r.Role, "apiName": r.ApiName, "createdBy": r.CreatedBy, "createdAt": r.CreatedAt, "updatedAt": r.UpdatedAt}
}

func (r *RBAC) MapData(data map[string]interface{}) {
	if val, ok := data["role"]; ok {
		r.Role = val.(string)
	}
	if val, ok := data["apiName"]; ok {
		r.ApiName = val.(string)
	}
	if val, ok := data["createdBy"]; ok {
		r.CreatedBy = val.(string)
	}
	if val, ok := data["createdAt"]; ok {
		r.CreatedAt = int64(val.(float64))
	}
}

func (r *RBAC) CleanUp() {
	r.ApiName = ""
	r.Role = ""
	r.CreatedBy = ""
	r.CreatedAt = 0
	r.UpdatedAt = 0
}
