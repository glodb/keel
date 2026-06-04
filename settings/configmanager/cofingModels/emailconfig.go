package configModels

type EmailConfig struct {
	SMPTClient         string `json:"smtpClient"`
	SMPTPort           string `json:"smtpPort"`
	Password           string `json:"password"`
	EmailAddress       string `json:"emailAddress"`
	EmailName          string `json:"emailName"`
	IsTLS              bool   `json:"isTLS"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}
