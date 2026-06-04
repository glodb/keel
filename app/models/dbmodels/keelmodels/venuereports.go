package keelmodels

type VenueReport struct {
	Id             string `json:"id" bson:"id"`
	VenueId        string `json:"venueId" bson:"venueId"`
	LastReportedAt int64  `json:"lastReportedAt" bson:"lastReportedAt"`
	ReportsCount   int    `json:"reportsCount" bson:"reportsCount"`
	IsActive       bool   `json:"isActive" bson:"isActive"`
	CreatedAt      int64  `json:"createdAt" bson:"createdAt"`
	ClearedAt      int64  `json:"clearedAt" bson:"clearedAt"`
	ClearedBy      string `json:"clearedBy" bson:"clearedBy"`
	ClearedReason  string `json:"clearedReason" bson:"clearedReason"`
}

type VenueReportDetail struct {
	Id                  string `json:"id" bson:"id"`
	ReportId            string `json:"reportId" bson:"reportId"`
	VenueId             string `json:"venueId" bson:"venueId"`
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
