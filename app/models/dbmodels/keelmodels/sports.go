package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// Sports represents a sports entity in the database
type Sports struct {
	SportsId            string `json:"sportsId" bson:"sportsId"`
	SportsName          string `json:"sportsName" bson:"sportsName"`
	SportsImage         string `json:"sportsImage" bson:"sportsImage"`
	ActiveGlobalPlayers int64  `json:"activeGlobalPlayers" bson:"activeGlobalPlayers"`
	SportSpot           int64  `json:"sportSpot" bson:"sportSpot"`
	CreatedAt           int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt           int64  `json:"updatedAt" bson:"updatedAt"`
}

// GetQuery returns the query for finding this sports record
func (s *Sports) GetQuery() map[string]interface{} {
	return map[string]interface{}{"sportsId": s.SportsId}
}

// GetUpdate returns the update data for this sports record
func (s *Sports) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"sportsName":  s.SportsName,
		"sportsImage": s.SportsImage,
		"createdAt":   time.Now().Unix(),
		"updatedAt":   time.Now().Unix(),
	}
}

// MapData maps the provided data to the Sports struct
func (s *Sports) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, s)
}

// EncodeRedisData encodes the Sports data for Redis storage
func (s *Sports) EncodeRedisData() string {
	bytes, _ := sonic.Marshal(s)
	return string(bytes)
}

// DecodeRedisData decodes the Sports data from Redis storage
func (s *Sports) DecodeRedisData(data string) error {
	return sonic.Unmarshal([]byte(data), s)
}

func (s *Sports) CleanUp() {
	s.SportsId = ""
	s.SportsName = ""
	s.SportsImage = ""
	s.ActiveGlobalPlayers = 0
	s.CreatedAt = 0
	s.UpdatedAt = 0
}
