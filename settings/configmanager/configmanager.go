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

// config is the single source of truth for all runtime configuration.
//
// Fields are loaded from two JSON files (global then service, service wins),
// then overridden by environment variables (see loadFromSystemEnv).
// Use NewConfig() + LoadFromBytes() + SetGlobalInstance() to supply config
// programmatically without any files on disk.
//
// Required fields are enforced by validateConfig() at startup and will produce
// error-level log lines but will NOT crash the process.
//
// Legend used in field comments:
//
//	[REQUIRED] — must be set; startup validation will report an error if absent
//	[OPTIONAL] — safe to omit; zero value is used or the feature is disabled
type config struct {

	// ── Core ─────────────────────────────────────────────────────────────────

	// ClassName is auto-set to the service name passed to Configure(). Do not
	// set this in JSON — it is overwritten on load.
	// [OPTIONAL] set programmatically
	ClassName string `json:"className"`

	// Address is the host:port the HTTP server listens on (e.g. ":8080").
	// [REQUIRED]
	Address string `json:"address"`

	// DeploymentEnv tags every Redis cache key and is used for environment-
	// aware behaviour. Must be one of: DEV, DEVELOPMENT, TEST, TESTING,
	// STAGING, PROD, PRODUCTION (case-insensitive).
	// Defaults to the env identifier passed to Configure() (e.g. "DEV").
	// Override via DEPLOYMENT_ENV env var.
	// [OPTIONAL — defaults to the Configure() env argument]
	DeploymentEnv string `json:"deploymentEnv"`

	// IsProduction / Production — either flag puts Gin in release mode.
	// [OPTIONAL — default false]
	IsProduction bool `json:"isProduction"`
	Production   bool `json:"production"`

	// PrintWarning / PrintInfo gate the Debug/Debugf log helpers on the
	// global logger.  Set PrintInfo=true in dev to see verbose output.
	// [OPTIONAL — default false]
	PrintWarning bool `json:"printWarning"`
	PrintInfo    bool `json:"printInfo"`

	// MicroServiceName is a human-readable label used in tracing / OpenAPI.
	// [OPTIONAL]
	MicroServiceName string `json:"microServiceName"`

	// ── Databases ────────────────────────────────────────────────────────────
	// At least one database block must be populated if you use the database
	// layer.  Fields within each block that are required when the block is
	// non-empty are documented in the configmodels package.

	// PSql holds PostgreSQL connection settings.
	// [OPTIONAL — omit if not using PostgreSQL]
	PSql configmodels.PSqlConfig `json:"psql"`

	// MySql holds MySQL connection settings.
	// [OPTIONAL — omit if not using MySQL]
	MySql configmodels.MySqlConfig `json:"mysql"`

	// Mongo holds MongoDB connection settings.
	// [OPTIONAL — omit if not using MongoDB]
	Mongo configmodels.MongoConfig `json:"mongo"`

	// ── Messaging / NATS pubsub ───────────────────────────────────────────────

	// NatsServerAddress is the NATS server URL (e.g. "nats://localhost:4222").
	// Override via NATS_SERVER env var.
	// [OPTIONAL — omit if not using NATS]
	NatsServerAddress string `json:"natsServer"`

	// RegisteredTopics lists NATS subjects this service creates/owns.
	// [OPTIONAL]
	RegisteredTopics []string `json:"registeredTopics"`

	// PublishingTopics lists NATS subjects this service publishes to.
	// [OPTIONAL]
	PublishingTopics []string `json:"publishingTopics"`

	// RpcRequestTopics lists NATS subjects used for outbound RPC requests.
	// [OPTIONAL]
	RpcRequestTopics []string `json:"rpcRequestTopics"`

	// SubscribedTopics maps NATS queue-group subjects → handler config.
	// [OPTIONAL]
	SubscribedTopics map[string]interface{} `json:"subscribedTopics"`

	// NonQueueSubscribedTopics is like SubscribedTopics but without queue
	// groups (every subscriber receives every message).
	// [OPTIONAL]
	NonQueueSubscribedTopics map[string]interface{} `json:"nonQueueSubscribedTopics"`

	// RpcSubscribedTopics maps NATS subjects → handler config for inbound RPC.
	// [OPTIONAL]
	RpcSubscribedTopics map[string]interface{} `json:"rpcSubscribedTopics"`

	// PublisherBatchSize is the max number of messages bundled in one NATS
	// publish call.  Must be ≤ 10 000.  0 disables batching.
	// [OPTIONAL — default 0]
	PublisherBatchSize int64 `json:"publisherBatchSize"`

	// ── Cache / Redis ─────────────────────────────────────────────────────────

	// CacheType selects the cache backend (currently only "redis" is supported).
	// [OPTIONAL — omit to disable caching]
	CacheType string `json:"cacheType"`

	// Redis holds the Redis connection pool settings.
	// [OPTIONAL — required only when CacheType = "redis"]
	Redis configmodels.RedisConnection `json:"redis"`

	// RedisRetries is the number of retries when acquiring a distributed lock
	// (total attempts = 1 + RedisRetries).  Must be ≤ 10.
	// [OPTIONAL — default 0 (one attempt, no retries)]
	RedisRetries int `json:"redisRetries"`

	// RedisRetryInterval is the sleep duration in milliseconds between lock
	// acquisition retries.
	// [OPTIONAL — default 0]
	RedisRetryInterval int `json:"redisRetryInterval"`

	// CacheContextTimeout is the max seconds a cache operation may block.
	// Must be ≤ 60.
	// [OPTIONAL — default 0 (no timeout enforced beyond the caller's context)]
	CacheContextTimeout int64 `json:"cacheContextTimeout"`

	// ── HTTP / routing ────────────────────────────────────────────────────────
	// These three fields control how route paths are assembled.
	// Final Gin path = ServiceLBName + ApiPrefix + "/" + Apis.ApiName

	// ServiceLBName is a path prefix prepended to every route, typically the
	// service's load-balancer path (e.g. "/api/v1").  May be empty.
	// [OPTIONAL]
	ServiceLBName string `json:"serviceLBName"`

	// ApiPrefix is a path segment inserted between ServiceLBName and each
	// individual route name (e.g. "/users").  May be empty.
	// [OPTIONAL]
	ApiPrefix string `json:"apiPrefix"`

	// PageSize is the default number of items returned in paginated responses.
	// Must be ≤ MaxPageSize when both are set.
	// [OPTIONAL — default 0]
	PageSize int `json:"pageSize"`

	// MaxPageSize caps the page size a client may request.  Must be ≤ 10 000.
	// [OPTIONAL — default 0 (no cap)]
	MaxPageSize int `json:"maxPageSize"`

	// SortKey is the query-parameter name used to specify sort order.
	// [OPTIONAL]
	SortKey string `json:"sortKey"`

	// TimeRangeKey is the query-parameter name used for time-range filtering.
	// [OPTIONAL]
	TimeRangeKey string `json:"timeRangeKey"`

	// ── RPC ───────────────────────────────────────────────────────────────────

	// RPCRequestExpirySeconds is how long an outbound RPC request waits for a
	// reply before timing out.  Must be ≤ 300.
	// [OPTIONAL — default 0]
	RPCRequestExpirySeconds int `json:"rpcRequestExpirySeconds"`

	// ── Secure cookies ────────────────────────────────────────────────────────

	// SecureCookieHash is the HMAC key for the securecookie package (32 bytes).
	// [OPTIONAL — required only if using settings/cookie]
	SecureCookieHash string `json:"secureCookieHash"`

	// SecureCookieBlock is the AES encryption key (16, 24, or 32 bytes).
	// [OPTIONAL — required only if using settings/cookie with encryption]
	SecureCookieBlock string `json:"secureCookieBlock"`

	// ── Soft-delete keys ──────────────────────────────────────────────────────

	// SoftDeleteCollectionPrefix is prepended to collection names for soft-
	// deleted record storage.
	// [OPTIONAL]
	SoftDeleteCollectionPrefix string `json:"softDeleteCollectionPrefix"`

	// SoftDeletionKey is the JSON field name written when a record is soft-deleted.
	// [OPTIONAL]
	SoftDeletionKey string `json:"softDeletionKey"`

	// DeletedByKey is the JSON field name recording who soft-deleted a record.
	// [OPTIONAL]
	DeletedByKey string `json:"deletedByKey"`

	// ── Migrations ────────────────────────────────────────────────────────────

	// MigrationsPath is the filesystem path to SQL migration files.
	// [OPTIONAL — required only if using database migrations]
	MigrationsPath string `json:"migrationsPath"`

	// ── Notifications ─────────────────────────────────────────────────────────

	// Email holds SMTP settings for email delivery.
	// [OPTIONAL — required only if sending email]
	Email configmodels.EmailConfig `json:"email"`

	// NotificationSender holds generic sender-channel configuration.
	// [OPTIONAL]
	NotificationSender configmodels.NotificationSenderConfig `json:"notificationSender"`

	// MessageSendingMilliSeconds is the minimum interval between notification
	// sends, in milliseconds.  Must be ≥ 0.
	// [OPTIONAL — default 0]
	MessageSendingMilliSeconds int `json:"messageSendingMilliSeconds"`

	// ── Firebase / FCM ────────────────────────────────────────────────────────

	// FirebaseCredentialsFileName is the path to the Firebase service-account
	// JSON key file (alternative to FirebaseCredentialsJson).
	// [OPTIONAL]
	FirebaseCredentialsFileName string `json:"firebaseCredentialsFileName"`

	// FirebaseMessageUrl is the FCM send endpoint URL.
	// [OPTIONAL]
	FirebaseMessageUrl string `json:"firebaseMessageUrl"`

	// FirebaseProjectId is the GCP project ID for the Firebase project.
	// [OPTIONAL]
	FirebaseProjectId string `json:"firebaseProjectId"`

	// FirebaseCredentialsJson is the inline JSON of the Firebase service-account
	// key (alternative to FirebaseCredentialsFileName).
	// [OPTIONAL]
	FirebaseCredentialsJson string `json:"firebaseCredentialsJson"`

	// ── Search ────────────────────────────────────────────────────────────────

	// MeilisearchConfig holds the Meilisearch host and API key.
	// [OPTIONAL — required only if using full-text search]
	MeilisearchConfig configmodels.MeilisearchConfig `json:"meilisearch"`

	// ── Observability ────────────────────────────────────────────────────────

	// MetricsPort is the address (e.g. ":9090") for the Prometheus metrics
	// HTTP server.  Empty disables the metrics server.
	// [OPTIONAL]
	MetricsPort string `json:"metricsPort"`

	// JaegerEndpoint is the Jaeger collector URL for OpenTelemetry traces.
	// Only used when UseTracing is true.
	// [OPTIONAL]
	JaegerEndpoint string `json:"jaegerEndpoint"`

	// UseTracing enables OpenTelemetry export to Jaeger. When false, tracing is
	// not initialized and no spans are exported.
	// [OPTIONAL — default false; set true or provide jaegerEndpoint to enable]
	UseTracing bool `json:"useTracing"`

	// UsePprof enables the Go pprof profiling HTTP endpoints.
	// [OPTIONAL — default false]
	UsePprof bool `json:"usePprof"`

	// PprofAddress is the listen address for the pprof server (e.g. ":6060").
	// [OPTIONAL — required when UsePprof = true]
	PprofAddress string `json:"pprofAddress"`

	// ── Socket.IO ────────────────────────────────────────────────────────────

	// SocketAddress is the listen address for the Socket.IO server
	// (e.g. ":9000").  Empty disables the socket server.
	// [OPTIONAL]
	SocketAddress string `json:"socketAddress"`

	// ── Versioning ───────────────────────────────────────────────────────────

	// ServiceVersion is embedded in OpenAPI specs and distributed traces.
	// [OPTIONAL]
	ServiceVersion genericmodels.Version `json:"serviceVersion"`

	// raw holds the decoded JSON (global merged with service) for consumer-only
	// config accessed via the typed GetX / GetXOr / GetXOK helpers below.
	raw map[string]interface{}
}

var getInstance = sync.OnceValue(func() *config {
	instance := &config{}
	instance.localSetup()
	return instance
})

// NewConfig returns a blank, ready-to-use config instance independent of the
// singleton. The raw map is initialized so Set() and GetX() helpers work
// immediately. Populate typed fields directly (e.g. cfg.Mongo.Host = "…")
// and/or call LoadFromBytes() to unmarshal JSON, then optionally call
// SetGlobalInstance(cfg) to make keel's own internals use it.
func NewConfig() *config {
	return &config{
		raw: make(map[string]interface{}),
	}
}

// globalOverride, when non-nil, is returned by GetInstance() instead of the
// lazy-loaded singleton. Set it with SetGlobalInstance() before any call to
// GetInstance() (typically before Boot()) so keel's own internals pick it up.
var (
	globalOverride   *config
	globalOverrideMu sync.RWMutex
)

// SetGlobalInstance replaces the config that GetInstance() returns.
// Call this before Boot() (or before the first GetInstance() call) to supply
// programmatically-built config without disk files. Pairs with NewConfig() and
// LoadFromBytes().
//
//	cfg := configmanager.NewConfig()
//	cfg.Address = ":8080"
//	cfg.Mongo.Host = "localhost"
//	configmanager.SetGlobalInstance(cfg)
//	keel.Boot("dev", "myservice")
func SetGlobalInstance(cfg *config) {
	globalOverrideMu.Lock()
	globalOverride = cfg
	globalOverrideMu.Unlock()
}

func GetInstance() *config {
	globalOverrideMu.RLock()
	if globalOverride != nil {
		defer globalOverrideMu.RUnlock()
		return globalOverride
	}
	globalOverrideMu.RUnlock()
	return getInstance()
}

// LoadFromBytes unmarshals global and service JSON config bytes into c without
// touching the filesystem.  Either argument may be nil/empty.  Service JSON
// takes precedence over global JSON on overlapping keys, exactly mirroring the
// file-based Setup() precedence rule.  Environment-variable overrides are
// applied afterwards, so they still win over everything.
//
//	cfg := configmanager.NewConfig()
//	cfg.LoadFromBytes(globalJSON, serviceJSON)
//	configmanager.SetGlobalInstance(cfg)
func (c *config) LoadFromBytes(globalJSON, serviceJSON []byte) error {
	if c.raw == nil {
		c.raw = make(map[string]interface{})
	}

	if len(globalJSON) > 0 {
		if err := sonic.Unmarshal(globalJSON, c); err != nil {
			return fmt.Errorf("configmanager: unmarshal global JSON: %w", err)
		}
		var raw map[string]interface{}
		if err := sonic.Unmarshal(globalJSON, &raw); err == nil {
			for k, v := range raw {
				c.raw[k] = v
			}
		}
	}

	if len(serviceJSON) > 0 {
		if err := sonic.Unmarshal(serviceJSON, c); err != nil {
			return fmt.Errorf("configmanager: unmarshal service JSON: %w", err)
		}
		var raw map[string]interface{}
		if err := sonic.Unmarshal(serviceJSON, &raw); err == nil {
			for k, v := range raw {
				c.raw[k] = v
			}
		}
	}

	// Environment variables take the highest precedence.
	c.loadSecretsFromEnv()
	c.finalizeTracingConfig()
	return nil
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
	}
	if envVal := os.Getenv("USE_TRACING"); envVal != "" {
		if useTracing, err := strconv.ParseBool(envVal); err == nil {
			c.UseTracing = useTracing
		}
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

// finalizeTracingConfig resolves whether tracing is enabled and applies defaults.
// Tracing is off unless useTracing is true in config/env, or a jaegerEndpoint is
// configured without an explicit useTracing:false. Set USE_TRACING=false to disable.
func (c *config) finalizeTracingConfig() {
	if _, explicit := c.raw["useTracing"]; explicit {
		if !c.UseTracing {
			return
		}
	} else if os.Getenv("USE_TRACING") == "false" {
		return
	} else if !c.UseTracing {
		if c.JaegerEndpoint != "" {
			c.UseTracing = true
		} else {
			return
		}
	}

	if c.UseTracing && c.JaegerEndpoint == "" {
		c.JaegerEndpoint = "http://localhost:14268/api/traces"
	}
}

func (c *config) Setup(configFileName, path, serviceName, workingDir string) {
	c.setup(configFileName, path, serviceName, workingDir)
}

func (c *config) localSetup() {
	configFileName, path, serviceName, workingDir := c.getConfigNameAndPath()
	c.setup(configFileName, path, serviceName, workingDir)
}

func (c *config) Set(key string, value interface{}) {
	c.raw[key] = value
}

// Setup loads the global and service config files, merges them, then applies
// environment variable overrides. The raw map is populated for consumer-only
// fields accessed via the typed GetX helpers.
func (c *config) setup(configFileName, path, serviceName, workingDir string) {
	serviceLowerName := strings.ToLower(serviceName)

	globalConfigPath := workingDir + "/config/" + configFileName + ".json"
	serviceConfigPath := fmt.Sprintf("%s/services/%s/config/%s.json", workingDir, serviceLowerName, configFileName)

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

	c.finalizeTracingConfig()

	// When no config file (or env var) set DeploymentEnv, fall back to the
	// environment identifier used to select the config file (e.g. "DEV", "PROD").
	// This keeps cache-key prefixes and other env-tagged identifiers consistent
	// even in test scenarios where no JSON file is present on disk.
	if c.DeploymentEnv == "" {
		c.DeploymentEnv = configuredEnv
	}

	// Validate configuration after loading.
	if err := c.validateConfig(); err != nil {
		logger.Log().Error("Configuration validation failed",
			logger.StringField("service", serviceName),
			logger.ErrorField("error", err))
	}

	logger.Log().Info("Configuration loaded successfully",
		logger.StringField("service", serviceName),
		logger.StringField("environment", configFileName),
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
