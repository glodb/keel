package configmanager

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/glodb/keel/models/genericmodels"
	configmodels "github.com/glodb/keel/settings/configmanager/configmodels"
	"github.com/glodb/keel/settings/logger"

	"github.com/bytedance/sonic"
)

var (
	configuredEnv     = "DEV"
	configuredService = "SSOSERVICE"
)

// Configure sets the deployment environment and service name before GetInstance() is called.
// It must be called before the first call to GetInstance(), typically at the start of Boot().
func Configure(env, service string) {
	configuredEnv = env
	configuredService = service
}

// Local config has preference over global config.
// Environment variables have preference over local config.

type config struct {
	// --- keel-internal fields ---

	ClassName      string `json:"className"`
	Address        string `json:"address"`
	DeploymentEnv  string `json:"deploymentEnv"`
	IsProduction   bool   `json:"isProduction"`
	Production     bool   `json:"production"`
	PrintWarning   bool   `json:"printWarning"`
	PrintInfo      bool   `json:"printInfo"`
	MicroServiceName string `json:"microServiceName"`

	// Databases
	PSql  configmodels.PSqlConfig  `json:"psql"`
	MySql configmodels.MySqlConfig `json:"mysql"`
	Mongo configmodels.MongoConfig `json:"mongo"`

	// Messaging / pubsub
	NatsServerAddress        string                 `json:"natsServer"`
	RegisteredTopics         []string               `json:"registeredTopics"`
	PublishingTopics         []string               `json:"publishingTopics"`
	RpcRequestTopics         []string               `json:"rpcRequestTopics"`
	SubscribedTopics         map[string]interface{} `json:"subscribedTopics"`
	NonQueueSubscribedTopics map[string]interface{} `json:"nonQueueSubscribedTopics"`
	RpcSubscribedTopics      map[string]interface{} `json:"rpcSubscribedTopics"`
	PublisherBatchSize       int64                  `json:"publisherBatchSize"`

	// Cache / Redis
	CacheType          string                       `json:"cacheType"`
	Redis              configmodels.RedisConnection `json:"redis"`
	RedisRetries       int                          `json:"redisRetries"`
	RedisRetryInterval int                          `json:"redisRetryInterval"`
	CacheContextTimeout int64                       `json:"cacheContextTimeout"`

	// HTTP / routing
	ServiceLBName string `json:"serviceLBName"`
	ApiPrefix        string `json:"apiPrefix"`
	PageSize         int    `json:"pageSize"`
	MaxPageSize      int    `json:"maxPageSize"`
	SortKey          string `json:"sortKey"`
	TimeRangeKey     string `json:"timeRangeKey"`

	// RPC
	RPCRequestExpirySeconds int `json:"rpcRequestExpirySeconds"`

	// Secure cookies (used by settings/cookie)
	SecureCookieHash  string `json:"secureCookieHash"`
	SecureCookieBlock string `json:"secureCookieBlock"`

	// Soft-delete keys (used by database drivers)
	SoftDeleteCollectionPrefix string `json:"softDeleteCollectionPrefix"`
	SoftDeletionKey            string `json:"softDeletionKey"`
	DeletedByKey               string `json:"deletedByKey"`

	// Migrations
	MigrationsPath string `json:"migrationsPath"`

	// Notifications
	Email               configmodels.EmailConfig              `json:"email"`
	NotificationSender  configmodels.NotificationSenderConfig `json:"notificationSender"`
	MessageSendingMilliSeconds int                            `json:"messageSendingMilliSeconds"`

	// Firebase (notification sender)
	FirebaseCredentialsFileName string `json:"firebaseCredentialsFileName"`
	FirebaseMessageUrl          string `json:"firebaseMessageUrl"`
	FirebaseProjectId           string `json:"firebaseProjectId"`
	FirebaseCredentialsJson     string `json:"firebaseCredentialsJson"`

	// Search
	MeilisearchConfig configmodels.MeilisearchConfig `json:"meilisearch"`

	// Observability
	MetricsPort    string `json:"metricsPort"`
	JaegerEndpoint string `json:"jaegerEndpoint"`
	UsePprof       bool   `json:"usePprof"`
	PprofAddress   string `json:"pprofAddress"`

	// Versioning (keel uses in tracing/openapi)
	ServiceVersion genericmodels.Version `json:"serviceVersion"`

	// raw holds the decoded JSON (global merged with service) for consumer-only
	// config accessed via the typed GetX / GetXOr / GetXOK helpers below.
	raw map[string]interface{}
}

var getInstance = sync.OnceValue(func() *config {
	instance := &config{}
	instance.Setup()
	return instance
})

func GetInstance() *config {
	return getInstance()
}

// loadSecretsFromEnv loads sensitive configuration from environment variables.
// Environment variables take precedence over JSON config values.
func (c *config) loadSecretsFromEnv() {
	c.loadFromSystemEnv()
}

// loadFromSystemEnv loads configuration from system environment variables.
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
	if envVal := os.Getenv("USE_PPROF"); envVal != "" {
		if usePprof, err := strconv.ParseBool(envVal); err == nil {
			c.UsePprof = usePprof
		}
	}
	if envVal := os.Getenv("PPROF_ADDRESS"); envVal != "" {
		c.PprofAddress = envVal
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
	if envVal := os.Getenv("FIREBASE_CREDENTIALS_JSON"); envVal != "" {
		c.FirebaseCredentialsJson = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_HASH"); envVal != "" {
		c.SecureCookieHash = envVal
	}
	if envVal := os.Getenv("SECURE_COOKIE_BLOCK"); envVal != "" {
		c.SecureCookieBlock = envVal
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
	} else if c.JaegerEndpoint == "" {
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
	if envVal := os.Getenv("MEILI_HOST"); envVal != "" {
		c.MeilisearchConfig.Host = envVal
	}
	if envVal := os.Getenv("MEILI_API_KEY"); envVal != "" {
		c.MeilisearchConfig.ApiKey = envVal
	}
}

// Setup loads the global and service config files, merges them, then applies
// environment variable overrides. The raw map is populated for consumer-only
// fields accessed via the typed GetX helpers.
func (c *config) Setup() {
	name, path, serviceName, workingDir := c.getConfigNameAndPath()
	serviceLowerName := strings.ToLower(serviceName)

	globalConfigPath := workingDir + "/config/" + name + ".json"
	serviceConfigPath := fmt.Sprintf("%s/services/%s/config/%s.json", workingDir, serviceLowerName, name)

	c.raw = make(map[string]interface{})

	// Load global config.
	globalBytes, err := os.ReadFile(globalConfigPath)
	if err != nil {
		logger.Log().Error("Error reading global config file",
			logger.StringField("file", globalConfigPath),
			logger.ErrorField("error", err),
		)
	} else {
		if err = sonic.Unmarshal(globalBytes, c); err != nil {
			logger.Log().Error("Error decoding global JSON",
				logger.StringField("file", globalConfigPath),
				logger.ErrorField("error", err),
			)
		}
		var globalRaw map[string]interface{}
		if err = sonic.Unmarshal(globalBytes, &globalRaw); err == nil {
			for k, v := range globalRaw {
				c.raw[k] = v
			}
		}
	}

	// Load service config (overrides global).
	serviceBytes, err := os.ReadFile(serviceConfigPath)
	if err != nil {
		logger.Log().Error("Error reading service config file",
			logger.StringField("file", serviceConfigPath),
			logger.ErrorField("error", err),
		)
	} else {
		if err = sonic.Unmarshal(serviceBytes, c); err != nil {
			logger.Log().Error("Error decoding service JSON",
				logger.StringField("file", serviceConfigPath),
				logger.ErrorField("error", err),
			)
		}
		var serviceRaw map[string]interface{}
		if err = sonic.Unmarshal(serviceBytes, &serviceRaw); err == nil {
			for k, v := range serviceRaw {
				c.raw[k] = v
			}
		}
	}

	c.ClassName = serviceName

	// Environment variable overrides (highest priority).
	c.loadSecretsFromEnv()

	// Validate configuration after loading.
	if err := c.validateConfig(); err != nil {
		logger.Log().Error("Configuration validation failed",
			logger.StringField("service", serviceName),
			logger.ErrorField("error", err))
	}

	logger.Log().Info("Configuration loaded successfully",
		logger.StringField("service", serviceName),
		logger.StringField("environment", name),
		logger.StringField("config_path", path),
	)
}

// getConfigNameAndPath resolves the config name and path from values set by Configure().
func (c *config) getConfigNameAndPath() (string, string, string, string) {
	conName := strings.ToLower(configuredEnv)
	serviceName := configuredService

	path, err := os.Getwd()
	if err != nil {
		logger.Log().Error("Error getting working directory",
			logger.ErrorField("error", err),
		)
		return conName, "", serviceName, path
	}
	conPath := path + "/config/" + conName + ".json"

	return conName, conPath, serviceName, path
}

func (c *config) GetPrintInfo() bool {
	return c.PrintInfo
}

func (c *config) SetVersion(version genericmodels.Version) {
	c.ServiceVersion = version
}

// GetCacheContextTimeout returns cache context timeout as time.Duration.
func (c *config) GetCacheContextTimeout() time.Duration {
	if c.CacheContextTimeout <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.CacheContextTimeout) * time.Second
}

// ---------------------------------------------------------------------------
// Typed getters for consumer-only config stored in the raw map.
// JSON numbers decode as float64; int getters convert accordingly.
// Three forms per type: GetX (zero value), GetXOr (default), GetXOK (presence).
// ---------------------------------------------------------------------------

func (c *config) GetStringOK(key string) (string, bool) {
	v, ok := c.raw[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func (c *config) GetStringOr(key, def string) string {
	if s, ok := c.GetStringOK(key); ok {
		return s
	}
	return def
}

func (c *config) GetString(key string) string { return c.GetStringOr(key, "") }

func (c *config) GetIntOK(key string) (int, bool) {
	v, ok := c.raw[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int(n), true
	case int:
		return n, true
	case int64:
		return int(n), true
	}
	return 0, false
}

func (c *config) GetIntOr(key string, def int) int {
	if n, ok := c.GetIntOK(key); ok {
		return n
	}
	return def
}

func (c *config) GetInt(key string) int { return c.GetIntOr(key, 0) }

func (c *config) GetInt64OK(key string) (int64, bool) {
	v, ok := c.raw[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return int64(n), true
	case int64:
		return n, true
	case int:
		return int64(n), true
	}
	return 0, false
}

func (c *config) GetInt64Or(key string, def int64) int64 {
	if n, ok := c.GetInt64OK(key); ok {
		return n
	}
	return def
}

func (c *config) GetInt64(key string) int64 { return c.GetInt64Or(key, 0) }

func (c *config) GetFloat64OK(key string) (float64, bool) {
	v, ok := c.raw[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

func (c *config) GetFloat64Or(key string, def float64) float64 {
	if f, ok := c.GetFloat64OK(key); ok {
		return f
	}
	return def
}

func (c *config) GetFloat64(key string) float64 { return c.GetFloat64Or(key, 0) }

func (c *config) GetBoolOK(key string) (bool, bool) {
	v, ok := c.raw[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func (c *config) GetBoolOr(key string, def bool) bool {
	if b, ok := c.GetBoolOK(key); ok {
		return b
	}
	return def
}

func (c *config) GetBool(key string) bool { return c.GetBoolOr(key, false) }

func (c *config) GetMapOK(key string) (map[string]interface{}, bool) {
	v, ok := c.raw[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]interface{})
	return m, ok
}

func (c *config) GetMapOr(key string, def map[string]interface{}) map[string]interface{} {
	if m, ok := c.GetMapOK(key); ok {
		return m
	}
	return def
}

func (c *config) GetMap(key string) map[string]interface{} { return c.GetMapOr(key, nil) }

func (c *config) GetArrayOK(key string) ([]interface{}, bool) {
	v, ok := c.raw[key]
	if !ok {
		return nil, false
	}
	a, ok := v.([]interface{})
	return a, ok
}

func (c *config) GetArrayOr(key string, def []interface{}) []interface{} {
	if a, ok := c.GetArrayOK(key); ok {
		return a
	}
	return def
}

func (c *config) GetArray(key string) []interface{} { return c.GetArrayOr(key, nil) }

func (c *config) GetStringArrayOK(key string) ([]string, bool) {
	a, ok := c.GetArrayOK(key)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(a))
	for _, v := range a {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out, true
}

func (c *config) GetStringArrayOr(key string, def []string) []string {
	if a, ok := c.GetStringArrayOK(key); ok {
		return a
	}
	return def
}

func (c *config) GetStringArray(key string) []string { return c.GetStringArrayOr(key, nil) }
