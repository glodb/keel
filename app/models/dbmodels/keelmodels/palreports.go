package keelmodels

type PalReport struct {
	Id             string `json:"id" bson:"id"`
	PalId          string `json:"palId" bson:"palId"`
	LastReportedAt int64  `json:"lastReportedAt" bson:"lastReportedAt"`
	ReportsCount   int    `json:"reportsCount" bson:"reportsCount"`
	IsActive       bool   `json:"isActive" bson:"isActive"`
	CreatedAt      int64  `json:"createdAt" bson:"createdAt"`
	ClearedAt      int64  `json:"clearedAt" bson:"clearedAt"`
	ClearedBy      string `json:"clearedBy" bson:"clearedBy"`
	ClearedReason  string `json:"clearedReason" bson:"clearedReason"`
}

type PalReportDetail struct {
	Id                  string `json:"id" bson:"id"`
	ReportId            string `json:"reportId" bson:"reportId"`
	PalId               string `json:"palId" bson:"palId"`
	ReportedAt          int64  `json:"reportedAt" bson:"reportedAt"`
	ReportedBy          string `json:"reportedBy" bson:"reportedBy"`
	ReportedReason      string `json:"reportedReason" bson:"reportedReason"`
	ReportedDescription string `json:"reportedDescription" bson:"reportedDescription"`
	ReportedStatus      string `json:"reportedStatus" bson:"reportedStatus"`
	ReportedType        string `json:"reportedType" bson:"reportedType"`
	ReportedSeverity    string `json:"reportedSeverity" bson:"reportedSeverity"`
	ReportedCategory    string `json:"reportedCategory" bson:"reportedCategory"`
	IsActive            bool   `json:"isActive" bson:"isActive"`
}
