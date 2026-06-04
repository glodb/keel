package jsonmodels

type AddLevelRequest struct {
	LevelName  string `json:"levelName" validate:"required" field:"levelName"`
	LevelImage string `json:"levelImage" validate:"required" field:"levelImage"`
}

type DeleteLevelRequest struct {
	LevelId string `json:"levelId" validate:"required" field:"levelId"`
}

type UpdateLevelImageRequest struct {
	LevelId    string `json:"levelId" validate:"required" field:"levelId"`
	LevelImage string `json:"levelImage" validate:"required" field:"levelImage"`
}
