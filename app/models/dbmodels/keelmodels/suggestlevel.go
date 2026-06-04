package keelmodels

type SuggestLevel struct {
	Id             string `json:"id" bson:"id"`
	UserId         string `json:"userId" bson:"userId"`
	SportId        string `json:"sportId" bson:"sportId"`
	CurrentLevelId string `json:"currentLevelId" bson:"currentLevelId"`
	SuggestedLevel string `json:"suggestedLevel" bson:"suggestedLevel"`
	SportName      string `json:"sportName" bson:"sportName"`
	LevelName      string `json:"levelName" bson:"levelName"`
	TargetUserId   string `json:"targetUserId" bson:"targetUserId"`
	CreatedAt      int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt      int64  `json:"updatedAt" bson:"updatedAt"`
}
