package configModels

type PaymentConfig struct {
	MaxConnections int `json:"maxConnections"`
}

type PayTabConfig struct {
	API         string `json:"api"`
	CallBackUrl string `json:"callBackUrl"`
	ClientKey   string `json:"clientKey"`
	ServerKey   string `json:"serverKey"`
	ProfileID   string `json:"profileId"`
	Currency    string `json:"currency"`
}

type StripeConfig struct {
	SecretKey string `json:"secretKey"`
	PublicKey string `json:"publicKey"`
	Currency  string `json:"currency"` // Default currency (e.g., "usd")
}
