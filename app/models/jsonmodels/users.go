package jsonmodels

import (
	"github.com/glodb/keel/app/models/dbmodels/keelmodels"
	"github.com/glodb/keel/settings/customtypes"
)

type SearchUsers struct {
	SearchPhrase string `json:"searchPhrase,omitempty" validate:"required"`
}

type UserLogin struct {
	CurrentNationalID string `json:"currentNationalId"`
	Phone             string `json:"phone"`
}

type VerifyLogin struct {
	OTP       string            `json:"otp" validate:"required" field:"otp"`
	SessionId string            `json:"sessionId,omitempty"`
	UserId    string            `json:"userId,omitempty"`
	Mode      customtypes.Modes `json:"mode,omitempty" `
}

type UserProfile struct {
	UserId            string        `json:"userId,omitempty"`
	Email             string        `json:"email,omitempty"`
	UserName          string        `json:"userName,omitempty"`
	Phone             string        `json:"phone,omitempty"`
	NationalIDs       []interface{} `json:"nationalIds,omitempty"`
	CurrentNationalID string        `json:"currentNationalId,omitempty"`
	AvatarUrl         string        `json:"avatarUrl,omitempty"`
}

type UpdateUser struct {
	UserId            string `json:"userId,omitempty"`
	UserName          string `json:"userName"`
	Email             string `json:"email"`
	CurrentNationalID string `json:"currentNationalId,omitempty"`
	AvatarUrl         string `json:"avatarUrl,omitempty"`
}

type NationalIdUpdate struct {
	NationalID             string `json:"nationalId" validate:"required"`
	IsSetCurrentNationalID bool   `json:"isSetCurrentNationalId"`
}

type UpdateProfile struct {
	Phone    string `json:"phone"`
	UserName string `json:"userName"`
	Email    string `json:"email"`
	Age      int    `json:"age"`
	Gender   string `json:"gender"`
	Weight   int    `json:"weight" `
	Religion string `json:"religion"`
}

type Profile struct {
	UserId            string        `json:"userId,omitempty"`
	UserName          string        `json:"userName,omitempty"`
	Phone             string        `json:"phone,omitempty"`
	Name              string        `json:"name,omitempty"`
	Email             string        `json:"email,omitempty"`
	NationalIDs       []interface{} `json:"nationalIds,omitempty"`
	CurrentNationalID string        `json:"currentNationalId,omitempty"`
	AvatarUrl         string        `json:"avatarUrl,omitempty"`
	KhRole            int           `json:"khRole,omitempty"`
	LastActivity      int64         `json:"lastActivity,omitempty"`
	App               string        `json:"app,omitempty"`
	Age               int           `json:"age,omitempty"`
	Gender            string        `json:"gender,omitempty"`
	Weight            float64       `json:"weight,omitempty"`
	Religion          string        `json:"religion,omitempty"`
	CreatedAt         int64         `json:"createdAt,omitempty"`
	UpdatedAt         int64         `json:"updatedAt,omitempty"`
	PhoneVerified     bool          `json:"phoneVerified,omitempty"`
}

type ProfileGet struct {
	UserId      string `json:"userId,omitempty" bson:"userId"`
	UserName    string `json:"userName,omitempty" bson:"userName"`
	Phone       string `json:"phone,omitempty" bson:"phone"`
	AvatarUrl   string `json:"avatarUrl,omitempty" bson:"avatarUrl"`
	IsFollowing bool   `json:"isFollowing,omitempty" bson:"isFollowing"`
}

type CheckPhoneNumber struct {
	Phone string `json:"phone" bson:"phone" validate:"required,e164" field:"phone"`
}

type CheckUserName struct {
	UserName string `json:"userName" bson:"userName" validate:"required,min=3,max=32,alphanum" field:"userName"`
}

type UpdateProfilePhoto struct {
	AvatarUrl string `json:"avatarUrl" bson:"avatarUrl" validate:"required,url" field:"avatarUrl"`
}

type LogoutUsers struct {
	PhoneNumbers []string `json:"phoneNumbers" validate:"required" field:"phoneNumbers"`
}

type UserCountRequest struct {
	StartTime int64 `json:"startTime" validate:"required" field:"startTime"`
	EndTime   int64 `json:"endTime" validate:"required" field:"endTime"`
}

type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email" field:"email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email" field:"email"`
	OTP   string `json:"otp" validate:"required" field:"otp"`
}

type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email" field:"email"`
	Password string `json:"password" validate:"required" field:"password"`
}

type ChangePasswordRequest struct {
	// OldPassword string `json:"oldPassword" validate:"required" field:"oldPassword"`
	NewPassword string `json:"newPassword" validate:"required,min=8,max=32,containsany=!@#$%^&*,excludesall=<>{}[]" field:"newPassword"`
}

type ForgetPasswordRequest struct {
	Email string `json:"email" validate:"required,email" field:"email"`
}

type ResetPasswordRequest struct {
	Email       string `json:"email" validate:"required,email" field:"email"`
	OTP         string `json:"otp" validate:"required" field:"otp"`
	NewPassword string `json:"newPassword" validate:"required,min=6" field:"newPassword"`
}

type RegisterUserRequest struct {
	Email    string `json:"email" validate:"required,email" field:"email"`
	Password string `json:"password" validate:"required,min=8,max=32,containsany=!@#$%^&*,excludesall=<>{}[]" field:"password"`
}

type IsEmailAvailableRequest struct {
	Email string `json:"email" validate:"required,email" field:"email"`
}

// SocialLoginRequest is used for Google/Facebook/Apple login
type SocialLoginRequest struct {
	IdToken  string `json:"idToken" validate:"required" field:"idToken"`   // The ID token from the provider
	Provider string `json:"provider" validate:"required" field:"provider"` // "google", "facebook", "apple"
	Email    string `json:"email,omitempty"`                               // Optional: email from provider
	Name     string `json:"name,omitempty"`                                // Optional: name from provider
	Avatar   string `json:"avatar,omitempty"`                              // Optional: avatar URL from provider
}

// GoogleUserInfo represents the response from Google's tokeninfo endpoint
type GoogleUserInfo struct {
	Sub           string `json:"sub"`            // Google user ID
	Email         string `json:"email"`          // User's email
	EmailVerified string `json:"email_verified"` // "true" or "false"
	Name          string `json:"name"`           // User's full name
	Picture       string `json:"picture"`        // Avatar URL
	GivenName     string `json:"given_name"`     // First name
	FamilyName    string `json:"family_name"`    // Last name
}

// FacebookUserInfo represents the response from Facebook's Graph API
type FacebookUserInfo struct {
	Id      string `json:"id"`    // Facebook user ID
	Email   string `json:"email"` // User's email
	Name    string `json:"name"`  // User's full name
	Picture struct {
		Data struct {
			Url string `json:"url"` // Avatar URL
		} `json:"data"`
	} `json:"picture"`
}

// AppleUserInfo represents the verified claims from Apple's identity token
type AppleUserInfo struct {
	Sub   string `json:"sub"`   // Apple user ID (unique, stable)
	Email string `json:"email"` // User's email (may be a relay address)
}

type RegisterUserWithPhoneRequest struct {
	Phone    string `json:"phone" validate:"required,e164" field:"phone"`
	Password string `json:"password" validate:"required,min=8,max=32,containsany=!@#$%^&*,excludesall=<>{}[]" field:"password"`
}

type UpdateUserRequest struct {
	Name                    string                              `json:"name,omitempty"`
	About                   string                              `json:"about,omitempty"`
	Language                string                              `json:"language,omitempty"`
	Address                 string                              `json:"address"`
	IsPrivate               bool                                `json:"isPrivate,omitempty"`
	City                    string                              `json:"city"`
	Country                 string                              `json:"country"`
	Phone                   string                              `json:"phone,omitempty"`
	PreferredCurrency       string                              `json:"preferredCurrency,omitempty"`
	NotificationPreferences *keelmodels.NotificationPreferences `json:"notificationPreferences,omitempty"`
}

// SearchUsersRequest represents a request to search for users
type SearchUsersRequest struct {
	SearchQuery string `json:"searchQuery"`
	Page        int    `json:"page"`
	Limit       int    `json:"limit"`
}

// PaginationInterface implementation for SearchUsersRequest
func (s *SearchUsersRequest) GetPage() int {
	if s.Page <= 0 {
		return 1
	}
	return s.Page
}

func (s *SearchUsersRequest) GetLimit() int {
	if s.Limit <= 0 {
		return 10
	}
	return s.Limit
}

func (s *SearchUsersRequest) GetSort() int {
	return -1
}

func (s *SearchUsersRequest) GetSortValue() string {
	return ""
}

func (s *SearchUsersRequest) GetRangeKey() string {
	return ""
}

func (s *SearchUsersRequest) GetStartTime() int64 {
	return 0
}

func (s *SearchUsersRequest) GetEndTime() int64 {
	return 0
}

func (s *SearchUsersRequest) Set(page, limit, sort int, sortValue, rangeKey string, startTime, endTime int64) {
	s.Page = page
	s.Limit = limit
}
