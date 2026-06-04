package configModels

type GoogleLoginConfig struct {
	ClientId     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	RedirectUrl  string   `json:"redirectUrl"`
	Scopes       []string `json:"scopes"`
	EndPoint     string   `json:"endPoint"`
}
