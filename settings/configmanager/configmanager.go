package configmanager

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/glodb/keel/app/models/genericmodels"
	configModels "github.com/glodb/keel/settings/configmanager/cofingModels"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utilsdatatypes"

	"github.com/bytedance/sonic"
)

//Local config has preference over global config
//Environment variables have preference over local config

type config struct {
	ClassName                   string
	Address                     string                                `json:"address"`
	PSql                        configModels.PSqlConfig               `json:"psql"`
	MySql                       configModels.MySqlConfig              `json:"mysql"`
	Mongo                       configModels.MongoConfig              `json:"mongo"`
	Email                       configModels.EmailConfig              `json:"email"`
	MeilisearchConfig           configModels.MeilisearchConfig        `json:"meilisearch"`
	AppleConfig                 configModels.AppleConfig              `json:"apple"`
	CacheType                   string                                `json:"cacheType"`
	ServiceLBName               string                                `json:"serviceLBName"`
	SessionKey                  string                                `json:"sessionKey"`
	RegisteredTopics            []string                              `json:"registeredTopics"`
	PublishingTopics            []string                              `json:"publishingTopics"`
	RpcRequestTopics            []string                              `json:"rpcRequestTopics"`
	PrintWarning                bool                                  `json:"printWarning"`
	PrintInfo                   bool                                  `json:"printInfo"`
	SubscribedTopics            map[string]interface{}                `json:"subscribedTopics"`
	NonQueueSubscribedTopics    map[string]interface{}                `json:"nonQueueSubscribedTopics"`
	RpcSubscribedTopics         map[string]interface{}                `json:"rpcSubscrtibedTopics"`
	MicroServiceName            string                                `json:"microServiceName"`
	PublisherBatchSize          int64                                 `json:"publisherBatchSize"`
	NatsServerAddress           string                                `json:"natsServer"`
	IsProduction                bool                                  `json:"isProduction"`
	SessionSecret               string                                `json:"sessionSecret"`
	Redis                       configModels.RedisConnection          `json:"redis"`
	ServiceLogName              string                                `json:"serviceLogName"`
	MapApis                     map[string][]string                   `json:"apis"`
	MapOpenApis                 map[string]map[string][]string        `json:"openApis"`
	MapAllowUnvalidatedApis     map[string][]string                   `json:"allowUnvalidatedApis"`
	TokenExpiry                 int64                                 `json:"tokenExpiry"`
	MapAcl                      map[string]map[string][]string        `json:"acl"`
	PageSize                    int                                   `json:"pageSize"`
	OtpExpirySeconds            int64                                 `json:"otpExpirySeconds"`
	UseSecureCookie             bool                                  `json:"useScureCookie"`
	OtpResendSeconds            int64                                 `json:"otpResendSeconds"`
	WriteError                  bool                                  `json:"writeError"`
	CookieName                  string                                `json:"cookieName"`
	CookieDomain                string                                `json:"cookieDomain"`
	CookiePath                  string                                `json:"cookiePath"`
	MongoControllers            []string                              `json:"mongoControllers"`
	MySqlControllers            []string                              `json:"mySqlControllers"`
	PSqlControllers             []string                              `json:"psqlControllers"`
	RedisRetries                int                                   `json:"redisRetries"`
	RedisRetryInterval          int                                   `json:"redisRetryInterval"`
	SetLock                     string                                `json:"setLock"`
	MainQueue                   string                                `json:"mainQueue"`
	DBBatchSize                 int                                   `json:"dbBatchSize"`
	RPCRequestExpirySeconds     int                                   `json:"rpcRequestExpirySeconds"`
	SecureCookieHash            string                                `json:"secureCookieHash"`
	SecureCookieBlock           string                                `json:"secureCookieBlock"`
	ActionsDirPath              string                                `json:"actionsDirPath"`
	MigrationsPath              string                                `json:"migrationsPath"`
	RunMigrations               bool                                  `json:"runMigrations"`
	SoftDeleteField             string                                `json:"softDeleteField"`
	SoftDeleteCollectionPrefix  string                                `json:"softDeleteCollectionPrefix"`
	SoftDeletionKey             string                                `json:"softDeletionKey"`
	DeletedByKey                string                                `json:"deletedByKey"`
	DeletedByPhoneKey           string                                `json:"deletedByPhoneKey"`
	EnabledNotifications        int                                   `json:"enableNotifications"`
	DeploymentEnv               string                                `json:"deploymentEnv"`
	NotificationSender          configModels.NotificationSenderConfig `json:"notificationSender"`
	Payment                     configModels.PaymentConfig            `json:"payment"`
	PayTab                      configModels.PayTabConfig             `json:"paytab"`
	Stripe                      *configModels.StripeConfig            `json:"stripe"`
	FirebaseCredentialsFileName string                                `json:"firebaseCredentialsFileName"`
	FirebaseMessageUrl          string                                `json:"firebaseMessageUrl"`
	FirebaseProjectId           string                                `json:"firebaseProjectId"`
	FirebaseCredentialsJson     string                                `json:"firebaseCredentialsJson"`
	AllUserNames                string                                `json:"allUserNames"`
	AllUserPhones               string                                `json:"allUserPhones"`
	AllUserEmails               string                                `json:"allUserEmails"`
	UnVerifiedUserEmails        string                                `json:"unVerifiedUserEmails"`
	UnVerifiedUserPhones        string                                `json:"unVerifiedUserPhones"`
	Version                     string                                `json:"version"`
	SortKey                     string                                `json:"sortKey"`
	TimeRangeKey                string                                `json:"timeRangeKey"`
	MessageSendingMilliSeconds  int                                   `json:"messageSendingMilliSeconds"`
	Production                  bool                                  `json:"production"`
	UsePprof                    bool                                  `json:"usePprof"`
	ReleaseMode                 bool                                  `json:"releaseMode"`
	EncryptionKey               string                                `json:"encryptionKey"`
	SecureCookieName            string                                `json:"secureCookieName"`
	SecureCookieSecret          string                                `json:"secureCookieSecret"`
	SecureSessionExpirySeconds  int64                                 `json:"secureSessionExpirySeconds"`
	CacheContextTimeout         int64                                 `json:"cacheContextTimeout"`
	RateLimit                   configModels.RateLimitConfig          `json:"rateLimit"`
	ServiceVersion              genericmodels.Version                 `json:"serviceVersion"`
	ServiceMinVersion           string                                `json:"serviceMinVersion"`
	ServiceMinVersionParsed     genericmodels.Version                 `json:"serviceMinVersionParsed"`
	MetricsPort                 string                                `json:"metricsPort"`
	LogLevel                    string                                `json:"logLevel"`
	JaegerEndpoint              string                                `json:"jaegerEndpoint"`
	DefaultRole                 string                                `json:"defaultRole"`
	ApiPrefix                   string                                `json:"apiPrefix"`
	UseGcpTracing               bool                                  `json:"useGcpTracing"`
	GcpProjectId                string                                `json:"gcpProjectId"`
	GcsBucketName               string                                `json:"gcsBucketName"`
	GcsCredentialsJson          string                                `json:"gcsCredentialsJson"`
	SupportCreditsPerMonth      int                                   `json:"supportCreditsPerMonth"`
	GeminiAPIKey                string                                `json:"geminiAPIKey"`
	Acl                         map[string]map[string]*utilsdatatypes.Set[string]
	Apis                        map[string]*utilsdatatypes.Set[string]
	OpenApis                    map[string]*utilsdatatypes.Set[string]
	AllowedUnvalidatedApis      map[string]*utilsdatatypes.Set[string]
	ApiRoles                    map[string]*utilsdatatypes.Set[string]
	PasswordTokenExpiry         int64
	MaxPageSize                 int `json:"maxPageSize"`
	// MapFolders holds raw folder entries from JSON ("folders": {"uploads": "...", ...}).
	// Global config provides the base set; service config entries override or extend it.
	// Use Folders (the merged result) at runtime via GetFolder().
	MapFolders map[string]string `json:"folders"`
	// Folders is the merged registry: global entries as base, service entries win on conflict.
	Folders map[string]string
}

var getInstance = sync.OnceValue(func() *config {
	instance := &config{}
	instance.OpenApis = make(map[string]*utilsdatatypes.Set[string])
	instance.AllowedUnvalidatedApis = make(map[string]*utilsdatatypes.Set[string])
	instance.ApiRoles = make(map[string]*utilsdatatypes.Set[string])
	instance.Setup()
	instance.Apis = make(map[string]*utilsdatatypes.Set[string])

	return instance
})

func GetInstance() *config {
	return getInstance()
}

// loadSecretsFromEnv loads sensitive configuration from environment variables
// Environment variables take precedence over JSON config values
func (c *config) loadSecretsFromEnv() {
	// Finally, load from system environment variables (highest priority)
	c.loadFromSystemEnv()
}

// loadFromSystemEnv loads configuration from system environment variables
func (c *config) loadFromSystemEnv() {

	if envVal := os.Getenv("PSQL_HOST"); envVal != "" {
		c.PSql.Host = envVal
	}
	if envVal := os.Getenv("PSQL_PORT"); envVal != "" {
		c.PSql.Port = envVal
	}
	if envVal := os.Getenv("PSQL_PASSWORD"); envVal != "" {
		c.PSql.Password = envVal
	}
	if envVal := os.Getenv("PSQL_USERNAME"); envVal != "" {
		c.PSql.Username = envVal
	}
	if envVal := os.Getenv("DEPLOYMENT_ENV"); envVal != "" {
		c.DeploymentEnv = envVal
	}

	if envVal := os.Getenv("MONGO_APP_NAME"); envVal != "" {
		c.Mongo.AppName = envVal
	}

	if envVal := os.Getenv("MONGO_ATLAS"); envVal != "" {
		if atlas, err := strconv.ParseBool(envVal); err == nil {
			c.Mongo.Atlas = atlas
		}
	}
	if envVal := os.Getenv("MONGO_HOST"); envVal != "" {
		c.Mongo.Host = envVal
	}
	if envVal := os.Getenv("MONGO_PORT"); envVal != "" {
		c.Mongo.Port = envVal
	}
	if envVal := os.Getenv("MONGO_DBNAME"); envVal != "" {
		c.Mongo.DBName = envVal
	}
	if envVal := os.Getenv("MONGO_SECURE_MONGO"); envVal != "" {
		if secureMongo, err := strconv.ParseBool(envVal); err == nil {
			c.Mongo.SecureMongo = secureMongo
		}
	}
	if envVal := os.Getenv("MONGO_CERT_FILE"); envVal != "" {
		c.Mongo.CertFile = envVal
	}
	if envVal := os.Getenv("MONGO_USERNAME"); envVal != "" {
		c.Mongo.Username = envVal
	}
	if envVal := os.Getenv("MONGO_PASSWORD"); envVal != "" {
		c.Mongo.Password = envVal
	}
	if envVal := os.Getenv("MONGO_MAX_CONNECTION"); envVal != "" {
		if maxConn, err := strconv.Atoi(envVal); err == nil {
			c.Mongo.MongoMaxConnections = maxConn
		}
	}
	if envVal := os.Getenv("SESSION_KEY"); envVal != "" {
		c.SessionKey = envVal
	}

	if envVal := os.Getenv("SESSION_SECRET"); envVal != "" {
		c.SessionSecret = envVal
	}

	if envVal := os.Getenv("NATS_SERVER"); envVal != "" {
		c.NatsServerAddress = envVal
	}
	if envVal := os.Getenv("REDIS_MAX_CONNECTIONS"); envVal != "" {
		if maxConn, err := strconv.Atoi(envVal); err == nil {
			c.Redis.RedisMaxConnections = maxConn
		}
	}
	if envVal := os.Getenv("REDIS_MAX_IDLE_CONNECTIONS"); envVal != "" {
		if maxIdleConn, err := strconv.Atoi(envVal); err == nil {
			c.Redis.RedisMaxIdleConnections = maxIdleConn
		}
	}
	if envVal := os.Getenv("REDIS_CON"); envVal != "" {
		c.Redis.RedisCon = envVal
	}
	if envVal := os.Getenv("REDIS_ADDRESS"); envVal != "" {
		c.Redis.RedisAddress = envVal
	}
	if envVal := os.Getenv("REDIS_PRINT_REDIS"); envVal != "" {
		if printRedis, err := strconv.ParseBool(envVal); err == nil {
			c.Redis.PrintRedis = printRedis
		}
	}
	if envVal := os.Getenv("REDIS_RETRIES"); envVal != "" {
		if retries, err := strconv.Atoi(envVal); err == nil {
			c.RedisRetries = retries
		}
	}
	if envVal := os.Getenv("REDIS_RETRY_INTERVAL"); envVal != "" {
		if retryInterval, err := strconv.Atoi(envVal); err == nil {
			c.RedisRetryInterval = retryInterval
		}
	}
	if envVal := os.Getenv("CACHE_TYPE"); envVal != "" {
		c.CacheType = envVal
	}
	if envVal := os.Getenv("USE_SECURE_COOKIE"); envVal != "" {
		if useSecureCookie, err := strconv.ParseBool(envVal); err == nil {
			c.UseSecureCookie = useSecureCookie
		}
	}
	if envVal := os.Getenv("USE_PPROF"); envVal != "" {
		if usePprof, err := strconv.ParseBool(envVal); err == nil {
			c.UsePprof = usePprof
		}
	}
	if envVal := os.Getenv("TOKEN_EXPIRY"); envVal != "" {
		if tokenExpiry, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			c.TokenExpiry = tokenExpiry
		}
	}
	if envVal := os.Getenv("MESSAGE_SENDING_MILLI"); envVal != "" {
		if msgSending, err := strconv.Atoi(envVal); err == nil {
			c.MessageSendingMilliSeconds = msgSending
		}
	}
	if envVal := os.Getenv("FIREBASE_PROJECT_ID"); envVal != "" {
		c.FirebaseProjectId = envVal
	}
	if envVal := os.Getenv("FIREBASE_CREDENTIALS_FILE"); envVal != "" {
		c.FirebaseCredentialsFileName = envVal
	}
	if envVal := os.Getenv("FIREBASE_MESSAGE_URL"); envVal != "" {
		c.FirebaseMessageUrl = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_HASH"); envVal != "" {
		c.SecureCookieHash = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_BLOCK"); envVal != "" {
		c.SecureCookieBlock = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_NAME"); envVal != "" {
		c.SecureCookieName = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_SECRET"); envVal != "" {
		c.SecureCookieSecret = envVal
	}
	if envVal := os.Getenv("SECURE_SESSION_EXPIRY_SECONDS"); envVal != "" {
		if secureSessionExpiry, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			c.SecureSessionExpirySeconds = secureSessionExpiry
		}
	}
	if envVal := os.Getenv("COOKIE_DOMAIN"); envVal != "" {
		c.CookieDomain = envVal
	}
	if envVal := os.Getenv("COOKIE_PATH"); envVal != "" {
		c.CookiePath = envVal
	}
	if envVal := os.Getenv("SORT_KEY"); envVal != "" {
		c.SortKey = envVal
	}
	if envVal := os.Getenv("TIME_RANGE_KEY"); envVal != "" {
		c.TimeRangeKey = envVal
	}
	if envVal := os.Getenv("CACHE_CONTEXT_TIMEOUT"); envVal != "" {
		if cacheTimeout, err := strconv.ParseInt(envVal, 10, 64); err == nil {
			c.CacheContextTimeout = cacheTimeout
		}
	}
	if envVal := os.Getenv("JAEGER_ENDPOINT"); envVal != "" {
		c.JaegerEndpoint = envVal
	} else {
		c.JaegerEndpoint = "http://localhost:14268/api/traces"
	}

	if envVal := os.Getenv("SMTP_CLIENT"); envVal != "" {
		c.Email.SMPTClient = envVal
	}
	if envVal := os.Getenv("SMTP_PORT"); envVal != "" {
		c.Email.SMPTPort = envVal
	}
	if envVal := os.Getenv("EMAIL_FROM"); envVal != "" {
		c.Email.EmailAddress = envVal
	}
	if envVal := os.Getenv("EMAIL_SENDER"); envVal != "" {
		c.Email.EmailName = envVal
	}
	if envVal := os.Getenv("EMAIL_PASSWORD"); envVal != "" {
		c.Email.Password = envVal
	}
	if envVal := os.Getenv("EMAIL_IS_TLS"); envVal != "" {
		if isTLS, err := strconv.ParseBool(envVal); err == nil {
			c.Email.IsTLS = isTLS
		}
	}
	if envVal := os.Getenv("EMAIL_INSECURE_SKIP_VERIFY"); envVal != "" {
		if insecureSkipVerify, err := strconv.ParseBool(envVal); err == nil {
			c.Email.InsecureSkipVerify = insecureSkipVerify
		}
	}
	if envVal := os.Getenv("USE_GCP_TRACING"); envVal != "" {
		if useGcpTracing, err := strconv.ParseBool(envVal); err == nil {
			c.UseGcpTracing = useGcpTracing
		}
	}
	if envVal := os.Getenv("GCP_PROJECT_ID"); envVal != "" {
		c.GcpProjectId = envVal
	}

	if envVal := os.Getenv("GCS_BUCKET_NAME"); envVal != "" {
		c.GcsBucketName = envVal
	}
	if envVal := os.Getenv("GCS_CREDENTIALS_JSON"); envVal != "" {
		c.GcsCredentialsJson = envVal
	}

	if envVal := os.Getenv("MEILI_HOST"); envVal != "" {
		c.MeilisearchConfig.Host = envVal
	}
	if envVal := os.Getenv("MEILI_API_KEY"); envVal != "" {
		c.MeilisearchConfig.ApiKey = envVal
	}

	if envVal := os.Getenv("FIREBASE_MESSAGE_URL"); envVal != "" {
		c.FirebaseMessageUrl = envVal
	}

	if envVal := os.Getenv("FIREBASE_CREDENTIALS_FILE"); envVal != "" {
		c.FirebaseCredentialsFileName = envVal
	}

	if envVal := os.Getenv("FIREBASE_PROJECT_ID"); envVal != "" {
		c.FirebaseProjectId = envVal
	}

	if envVal := os.Getenv("FIREBASE_CREDENTIALS_JSON"); envVal != "" {
		c.FirebaseCredentialsJson = envVal
	}

	c.Stripe = &configModels.StripeConfig{}
	if envVal := os.Getenv("STRIPE_SECRET_KEY"); envVal != "" {
		c.Stripe.SecretKey = envVal
	}
	if envVal := os.Getenv("STRIPE_PUBLIC_KEY"); envVal != "" {
		c.Stripe.PublicKey = envVal
	}
	if envVal := os.Getenv("GEMINI_API_KEY"); envVal != "" {
		c.GeminiAPIKey = envVal
	}
	if envVal := os.Getenv("APPLE_TEAM_ID"); envVal != "" {
		c.AppleConfig.TeamId = envVal
	}
	if envVal := os.Getenv("APPLE_KEY_ID"); envVal != "" {
		c.AppleConfig.KeyId = envVal
	}
	if envVal := os.Getenv("APPLE_BUNDLE_ID"); envVal != "" {
		c.AppleConfig.BundleId = envVal
	}
	if envVal := os.Getenv("APPLE_SERVICE_ID"); envVal != "" {
		c.AppleConfig.ServiceId = envVal
	}
	if envVal := os.Getenv("APPLE_SERVICE_FILE_URL"); envVal != "" {
		c.AppleConfig.ServiceFileUrl = envVal
	}
	if envVal := os.Getenv("APPLE_JWKS_URL"); envVal != "" {
		c.AppleConfig.JWKSUrl = envVal
	}
}

// Setup The local config has preference over global config
func (c *config) Setup() {
	name, path, serviceName, workingDir := c.getConfigNameAndPath()

	serviceLowerName := strings.ToLower(serviceName)
	globalConfigPath := workingDir + "/config/" + name + ".json"
	globalConfigFile, err := os.Open(globalConfigPath)

	if err != nil {
		logger.Log().Error("Error opening global config file",
			logger.StringField("file", globalConfigPath),
			logger.ErrorField("error", err),
		)
		return
	}
	defer globalConfigFile.Close()

	decoder := sonic.ConfigDefault.NewDecoder(globalConfigFile)
	err = decoder.Decode(&c)

	if err != nil {
		logger.Log().Error("Error decoding global JSON",
			logger.StringField("file", globalConfigPath),
			logger.ErrorField("error", err),
		)
		return
	}

	// Snapshot global folders before service config overwrites MapFolders.
	globalFolders := make(map[string]string, len(c.MapFolders))
	for k, v := range c.MapFolders {
		globalFolders[k] = v
	}

	configFile, err := os.Open(fmt.Sprintf("%s/services/%s/config/%s.json", workingDir, serviceLowerName, name))
	if err != nil {
		logger.Log().Error("Error opening config file",
			logger.StringField("file", path+name+".json"),
			logger.ErrorField("error", err),
		)
		return
	}
	defer configFile.Close()

	decoder = sonic.ConfigDefault.NewDecoder(configFile)
	err = decoder.Decode(&c)

	if err != nil {
		logger.Log().Error("Error decoding JSON",
			logger.StringField("file", path+name+".json"),
			logger.ErrorField("error", err),
		)
		return
	}

	// Merge folder registry: global is the base, service entries override or extend.
	merged := make(map[string]string, len(globalFolders)+len(c.MapFolders))
	for k, v := range globalFolders {
		merged[k] = v
	}
	for k, v := range c.MapFolders {
		merged[k] = v
	}
	c.Folders = merged

	c.ClassName = serviceName

	// Load secrets from environment variables (overrides JSON config)
	c.loadSecretsFromEnv()

	c.Acl = make(map[string]map[string]*utilsdatatypes.Set[string])
	for k, v := range c.MapAcl {
		rawSet := utilsdatatypes.NewSet[string]()
		c.Acl[k] = make(map[string]*utilsdatatypes.Set[string])
		for innerK, innerV := range v {
			for _, val := range innerV {
				rawSet.Add(val)
			}
			c.Acl[k][innerK] = rawSet
		}
	}

	c.Apis = make(map[string]*utilsdatatypes.Set[string])
	c.OpenApis = make(map[string]*utilsdatatypes.Set[string])
	c.AllowedUnvalidatedApis = make(map[string]*utilsdatatypes.Set[string])
	c.ApiRoles = make(map[string]*utilsdatatypes.Set[string])

	for k, v := range c.MapApis {
		rawSet := utilsdatatypes.NewSet[string]()
		for _, val := range v {
			rawSet.Add(val)
		}
		c.Apis[k] = rawSet
	}

	for k, v := range c.MapOpenApis {

		if v["roles"] != nil {
			c.ApiRoles[k] = utilsdatatypes.NewSet[string]()
			for _, val := range v["roles"] {
				c.ApiRoles[k].Add(val)
			}
		}

		if v["methods"] != nil {
			rawSet := utilsdatatypes.NewSet[string]()
			for _, val := range v["methods"] {
				rawSet.Add(val)
			}
			c.OpenApis[k] = rawSet
		}
	}

	for k, v := range c.MapAllowUnvalidatedApis {
		rawSet := utilsdatatypes.NewSet[string]()
		for _, val := range v {
			rawSet.Add(val)
		}
		c.AllowedUnvalidatedApis[k] = rawSet
	}

	// Validate configuration after loading
	if err := c.validateConfig(); err != nil {
		logger.Log().Error("Configuration validation failed",
			logger.StringField("service", serviceName),
			logger.ErrorField("error", err))
		// Don't return error, just log warnings for now
		// In production, you might want to fail fast on critical errors
	}

	c.RateLimit.Init()
	c.RateLimit.Window, err = time.ParseDuration(c.RateLimit.WindowData)

	if err != nil {
		c.RateLimit.Window = 60 * time.Second
	}

	var version genericmodels.Version

	if c.ServiceMinVersion == "" {
		c.ServiceMinVersionParsed = genericmodels.Version{Major: 1, Minor: 0, Patch: 0}
	} else {
		err = version.Parse(c.ServiceMinVersion)
		if err != nil {
			c.ServiceMinVersionParsed = genericmodels.Version{Major: 1, Minor: 0, Patch: 0}
		} else {
			c.ServiceMinVersionParsed = version
		}
	}

	logger.Log().Info("Configuration loaded successfully",
		logger.StringField("service", serviceName),
		logger.StringField("environment", name),
		logger.StringField("config_path", path),
	)
}

// GetConfigNameAndPath get the config name on the basis of flag
func (c *config) getConfigNameAndPath() (string, string, string, string) {
	serverType := flag.String("env", "DEV", "use development server by default")
	configPath := flag.String("con", "SSOSERVICE", "use Uploader server by default")

	var conName string
	var conPath string
	flag.Parse()
	conName = strings.ToLower(*serverType)
	path, err := os.Getwd()
	if err != nil {
		logger.Log().Error("Error getting working directory",
			logger.ErrorField("error", err),
		)
		return conName, conPath, *configPath, path
	}
	conPath = path + "/config/" + conName + ".json"

	return conName, conPath, *configPath, path
}

func (c *config) GetPrintInfo() bool {
	return c.PrintInfo
}

func (c *config) SetVersion(version genericmodels.Version) {
	c.ServiceVersion = version
}

// GetFolder returns the folder path registered under the given name.
// Service-level config takes priority over global config.
// Returns ("", false) when the name is not registered.
func (c *config) GetFolder(name string) (string, bool) {
	if c.Folders == nil {
		return "", false
	}
	v, ok := c.Folders[name]
	return v, ok
}
