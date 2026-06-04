package jsonmodels

// AddSportsRequest represents the request to add a new sport
type AddSportsRequest struct {
	SportsName          string `json:"sportsName" validate:"required" field:"sportsName"`
	SportsImage         string `json:"sportsImage" validate:"required" field:"sportsImage"`
	SportSpot           int    `json:"sportSpot" validate:"required" field:"sportSpot"`
	ActiveGlobalPlayers int    `json:"activeGlobalPlayers" validate:"required" field:"activeGlobalPlayers"`
	CreatedAt           int64  `json:"createdAt" field:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt" field:"updatedAt"`
}

// DeleteSportsRequest represents the request to delete a sport
type DeleteSportsRequest struct {
	SportsId string `json:"sportsId" validate:"required" field:"sportsId"`
}

// UpdateSportsImageRequest represents the request to update a sport's image
type UpdateSportsImageRequest struct {
	SportsId    string `json:"sportsId" validate:"required" field:"sportsId"`
	SportsImage string `json:"sportsImage" validate:"required" field:"sportsImage"`
}
