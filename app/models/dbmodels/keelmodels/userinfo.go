package keelmodels

import (
	"time"

	"github.com/bytedance/sonic"
)

type UserSports struct {
	SportId string `json:"sportId" bson:"sportId"`
	LevelId string `json:"levelId" bson:"levelId"`
}

type UserSportsFilter struct {
	UserId    string `json:"userId" bson:"userId"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email" bson:"email"`
	AvatarUrl string `json:"avatarUrl" bson:"avatarUrl"`
	SportId   string `json:"sportId" bson:"sportId"`
	LevelId   string `json:"levelId" bson:"levelId"`
	IsPrivate bool   `json:"isPrivate" bson:"isPrivate"`
	CreatedAt int64  `json:"createdAt" bson:"createdAt"`
	UpdatedAt int64  `json:"updatedAt" bson:"updatedAt"`
}

type UserInfo struct {
	SessionId           string          `json:"sessionId"`
	UserId              string          `json:"userId" bson:"userId"`
	Name                string          `json:"name" bson:"name"`
	Email               string          `json:"email" bson:"email"`
	UserName            string          `json:"userName" bson:"userName"`
	Phone               string          `json:"phone" bson:"phone"`
	Password            string          `json:"password,omitempty"  bson:"password"`
	IsVerified          bool            `json:"isVerified" bson:"isVerified"`
	About               string          `json:"about" bson:"about"`
	SportsAdded         bool            `json:"sportsAdded" bson:"sportsAdded"`
	LocationAdded       bool            `json:"locationAdded" bson:"locationAdded"`
	ScheduleAdded       bool            `json:"scheduleAdded" bson:"scheduleAdded"`
	NationalIDs         []interface{}   `json:"nationalIds" bson:"nationalIds"`
	CurrentNationalID   string          `json:"currentNationalId" bson:"currentNationalId"`
	AllowedApps         []interface{}   `json:"allowedApps" bson:"allowedApps"`
	IsBlocked           bool            `json:"isBlocked" bson:"isBlocked"`
	UpdateSession       bool            `json:"updateSession" bson:"updateSession"`
	FromApp             string          `json:"fromApp" bson:"fromApp"`
	CreatedAt           int64           `json:"createdAt" bson:"createdAt"`
	UpdatedAt           int64           `json:"updatedAt" bson:"updatedAt"`
	TotalUnreadCount    int             `json:"totalUnreadCount" bson:"totalUnreadCount"`
	SkipTokenGeneration bool            `json:"skipTokenGeneration"`
	AvatarUrl           string          `json:"avatarUrl" bson:"avatarUrl"`
	UserSports          []UserSports    `json:"userSports" bson:"userSports"`
	UserLat             float64         `json:"userLat" bson:"userLat"`
	UserLng             float64         `json:"userLng" bson:"userLng"`
	UserAddress         string          `json:"userAddress" bson:"userAddress"`
	UserCity            string          `json:"userCity" bson:"userCity"`
	UserCountry         string          `json:"userCountry" bson:"userCountry"`
	UserRadius          int             `json:"userRadius" bson:"userRadius"`
	UserSchedule        *WeeklySchedule `json:"userSchedule" bson:"userSchedule"`
	DailySchedules      []DailySchedule `json:"dailySchedules" bson:"dailySchedules"`
	IsRepeatable        bool            `json:"isRepeatable" bson:"isRepeatable"`
	Role                string          `json:"role" bson:"role"`
	IsPrivate           bool            `json:"isPrivate" bson:"isPrivate"`
	Language            string          `json:"language" bson:"language"`
	IsSocialLogin       bool            `json:"isSocialLogin" bson:"isSocialLogin"`
	AuthProvider        string          `json:"authProvider,omitempty" bson:"authProvider,omitempty"` // "email", "google", "facebook", "apple"
	ProviderId          string          `json:"providerId,omitempty" bson:"providerId,omitempty"`     // Provider's user ID
	// Payments
	StripeCustomerId        string                  `json:"stripeCustomerId,omitempty" bson:"stripeCustomerId,omitempty"`
	PreferredCurrency       string                  `json:"preferredCurrency,omitempty" bson:"preferredCurrency,omitempty"`
	NotificationPreferences NotificationPreferences `json:"notificationPreferences,omitempty" bson:"notificationPreferences,omitempty"`
}

type NotificationPreferences struct {
	Email bool `json:"email" bson:"email"`
	Push  bool `json:"push" bson:"push"`
}

type PalInfo struct {
	UserId    string `json:"userId" bson:"userId"`
	Name      string `json:"name" bson:"name"`
	Email     string `json:"email" bson:"email"`
	AvatarUrl string `json:"avatarUrl" bson:"avatarUrl"`
}
type User struct {
	Token string `json:"id"`
}

func (ui *UserInfo) GetQuery() map[string]interface{} {
	return map[string]interface{}{"userId": ui.UserId}
}

func (ui *UserInfo) GetUpdate() map[string]interface{} {
	return map[string]interface{}{"email": ui.Email, "userName": ui.UserName, "name": ui.Name,
		"phone": ui.Phone, "password": ui.Password, "nationalIds": []interface{}{ui.CurrentNationalID},
		"currentNationalId": ui.CurrentNationalID, "createdAt": time.Now().Unix(),
		"allowedApps": ui.AllowedApps, "fromApp": ui.FromApp,
		"isBlocked": ui.IsBlocked, "updatedAt": time.Now().Unix()}
}

func (ui *UserInfo) MapData(data map[string]interface{}) {

	if val, ok := data["userId"]; ok {
		ui.UserId = val.(string)
	}

	if val, ok := data["email"]; ok {
		ui.Email = val.(string)
	}

	if val, ok := data["userName"]; ok {
		ui.UserName = val.(string)
	}

	if val, ok := data["phone"]; ok {
		ui.Phone = val.(string)
	}

	if val, ok := data["password"]; ok {
		ui.Password = val.(string)
	}
	if val, ok := data["currentNationalId"]; ok {
		ui.CurrentNationalID = val.(string)
	}

	if val, ok := data["allowedApps"]; ok {
		ui.AllowedApps = val.([]interface{})
	}

	if val, ok := data["isBlocked"]; ok {
		ui.IsBlocked = val.(bool)
	}

	if val, ok := data["name"]; ok {

		if val, ok := data["about"]; ok {
			ui.About = val.(string)
		}
		ui.Name = val.(string)
	}
	// if val, ok := data["fromApp"]; ok {
	// 	ui.FromApp = val.(string)
	// }

}

func (ui *UserInfo) CleanUp() {
	ui.UserId = ""
	ui.Email = ""
	ui.UserName = ""
	ui.Phone = ""
	ui.Password = ""
	ui.CurrentNationalID = ""
	ui.FromApp = ""
	ui.CreatedAt = 0
	ui.UpdatedAt = 0
	ui.Name = ""
	ui.About = ""
}

func (ts *UserInfo) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *UserInfo) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
