package configModels

type AppleConfig struct {
	TeamId         string `json:"teamId"`
	KeyId          string `json:"keyId"`
	BundleId       string `json:"bundleId"`
	ServiceId      string `json:"serviceId"`
	ServiceFileUrl string `json:"serviceFileUrl"`
	JWKSUrl        string `json:"jwksUrl"` // overridable via APPLE_JWKS_URL env (for local testing)
}
