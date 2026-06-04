package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

// Levels represents a levels entity in the database
type Levels struct {
	LevelId    string `json:"levelId" bson:"levelId"`
	LevelName  string `json:"levelName" bson:"levelName"`
	LevelImage string `json:"levelImage" bson:"levelImage"`
	CreatedAt  int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt" bson:"updatedAt"`
}

// GetQuery returns the query for finding this sports record
func (s *Levels) GetQuery() map[string]interface{} {
	return map[string]interface{}{"levelId": s.LevelId}
}

// GetUpdate returns the update data for this sports record
func (s *Levels) GetUpdate() map[string]interface{} {
	return map[string]interface{}{
		"levelName":  s.LevelName,
		"levelImage": s.LevelImage,
		"createdAt":  time.Now().Unix(),
		"updatedAt":  time.Now().Unix(),
	}
}

// MapData maps the provided data to the Sports struct
func (s *Levels) MapData(data map[string]interface{}) {
	bytes, _ := sonic.Marshal(data)
	sonic.Unmarshal(bytes, s)
}

// EncodeRedisData encodes the Sports data for Redis storage
func (s *Levels) EncodeRedisData() string {
	bytes, _ := sonic.Marshal(s)
	return string(bytes)
}

// DecodeRedisData decodes the Sports data from Redis storage
func (s *Levels) DecodeRedisData(data string) error {
	return sonic.Unmarshal([]byte(data), s)
}

func (s *Levels) CleanUp() {
	s.LevelId = ""
	s.LevelName = ""
	s.LevelImage = ""
	s.CreatedAt = 0
	s.UpdatedAt = 0
}
