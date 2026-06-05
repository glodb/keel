package baseconnections

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/glodb/keel/database/basetypes"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	readPreference        = "secondaryPreferred"
	connectionStringAtlas = "mongodb+srv://%s:%s@%s/%s?retryWrites=true&w=majority&appName=%s"
	connectionStringMain  = "mongodb://%s:%s@%s/%s?tls=true&replicaSet=rs0&readpreference=%s&retryWrites=false"
	connectionStringDev   = "mongodb://%s/%s"
	connectTimeout        = 30
)

type MongoConnection struct {
	dbName string
	client *mongo.Client
	params *ConnectionParams
}

func (u *MongoConnection) getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("failed parsing pem file")
	}

	return tlsConfig, nil
}

func (u *MongoConnection) CreateConnection() (ConnectionInterface, error) {
	var (
		host, port, username, password, dbName string
		atlas, secureMongo                     bool
		certFile, appName                      string
	)

	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
		username = u.params.Username
		password = u.params.Password
		dbName = u.params.DBName
		atlas = u.params.Atlas
		secureMongo = u.params.SecureMongo
		certFile = u.params.CertFile
		appName = u.params.AppName
	} else {
		cfg := configmanager.GetInstance().Mongo
		host = cfg.Host
		port = cfg.Port
		username = cfg.Username
		password = cfg.Password
		dbName = cfg.DBName
		atlas = cfg.Atlas
		secureMongo = cfg.SecureMongo
		certFile = cfg.CertFile
		appName = cfg.AppName
	}

	logger.Log().Info("Creating MongoDB connection",
		logger.StringField("username", username),
		logger.StringField("host", host),
		logger.StringField("port", port),
		logger.StringField("dbName", dbName),
		logger.BoolField("atlas", atlas),
		logger.BoolField("secureMongo", secureMongo),
		logger.StringField("certFile", certFile),
		logger.StringField("appName", appName))

	// Base client options applied to every connection mode.
	// RetryReads/RetryWrites: driver retries once on stale connections (e.g. Docker closing idle TCP).
	// MaxConnIdleTime: evict idle connections from the pool before the network layer kills them.
	baseOpts := options.Client().
		SetRetryReads(true).
		SetRetryWrites(true).
		SetMaxConnIdleTime(30 * time.Second)

	var err error
	var client *mongo.Client

	if atlas {
		connectionURI := fmt.Sprintf(connectionStringAtlas, username, password, host, dbName, appName)
		client, err = mongo.NewClient(baseOpts.ApplyURI(connectionURI))
		if err != nil {
			return nil, errors.New("failed to create client")
		}
	} else if secureMongo {
		connectionURI := fmt.Sprintf(connectionStringMain, username, password, host+":"+port, dbName, readPreference)
		tlsConfig, err := u.getCustomTLSConfig(certFile)
		if err != nil {
			return nil, errors.New("unable to get tls config")
		}
		client, err = mongo.NewClient(baseOpts.ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
		if err != nil {
			return nil, errors.New("failed to create client")
		}
	} else {
		connectionURI := fmt.Sprintf(connectionStringDev, host+":"+port, dbName)
		client, err = mongo.NewClient(baseOpts.ApplyURI(connectionURI))
		if err != nil {
			return nil, errors.New("failed to create client")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		logger.Log().Error("Failed to connect to cluster", logger.ErrorField("error", err))
		return nil, err
	}

	// Force a connection to verify our connection string
	err = client.Ping(ctx, nil)
	if err != nil {
		logger.Log().Error("Failed to ping cluster", logger.ErrorField("error", err))
		return nil, err
	}
	u.client = client
	u.dbName = dbName
	return u, nil
}

func (u *MongoConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.client
}

func (u *MongoConnection) Close() error {
	if u.client != nil {
		return u.client.Disconnect(context.Background())
	}
	return nil
}

func (u *MongoConnection) Ping(ctx context.Context) error {
	if u.client != nil {
		return u.client.Ping(ctx, nil)
	}
	return errors.New("client is nil")
}

func (u *MongoConnection) GetConnectionInfo() ConnectionInfo {
	var host, port string
	if u.params != nil {
		host = u.params.Host
		port = u.params.Port
	} else {
		cfg := configmanager.GetInstance().Mongo
		host = cfg.Host
		port = cfg.Port
	}
	return ConnectionInfo{
		DatabaseType: basetypes.MONGO,
		Host:         host,
		Port:         port,
		DatabaseName: u.dbName,
		ConnectionState: func() string {
			if u.client != nil {
				return "connected"
			}
			return "disconnected"
		}(),
	}
}

func (u *MongoConnection) IsHealthy(ctx context.Context) (bool, error) {
	if u.client == nil {
		return false, errors.New("client is nil")
	}

	err := u.client.Ping(ctx, nil)
	return err == nil, err
}
