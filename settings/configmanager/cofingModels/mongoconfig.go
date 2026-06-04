package configModels

type MongoConfig struct {
	Host                string `json:"host"`
	Port                string `json:"port"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	DBName              string `json:"dbname"`
	Atlas               bool   `json:"atlas"`
	SecureMongo         bool   `json:"secureMongo"`
	CertFile            string `json:"certFile"`
	MongoMaxConnections int    `json:"mongoMaxConnections"`
	AppName             string `json:"appName"`
}
