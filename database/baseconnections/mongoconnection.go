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
	var err error
	var client *mongo.Client

	logger.Log().Info("Creating MongoDB connection",
		logger.StringField("username", configmanager.GetInstance().Mongo.Username),
		logger.StringField("password", configmanager.GetInstance().Mongo.Password),
		logger.StringField("host", configmanager.GetInstance().Mongo.Host),
		logger.StringField("port", configmanager.GetInstance().Mongo.Port),
		logger.StringField("dbName", configmanager.GetInstance().Mongo.DBName),
		logger.BoolField("atlas", configmanager.GetInstance().Mongo.Atlas),
		logger.BoolField("secureMongo", configmanager.GetInstance().Mongo.SecureMongo),
		logger.StringField("certFile", configmanager.GetInstance().Mongo.CertFile),
		logger.IntField("mongoMaxConnections", configmanager.GetInstance().Mongo.MongoMaxConnections),
		logger.StringField("appName", configmanager.GetInstance().Mongo.AppName))

	// Base client options applied to every connection mode.
	// RetryReads/RetryWrites: driver retries once on stale connections (e.g. Docker closing idle TCP).
	// MaxConnIdleTime: evict idle connections from the pool before the network layer kills them.
	baseOpts := options.Client().
		SetRetryReads(true).
		SetRetryWrites(true).
		SetMaxConnIdleTime(30 * time.Second)

	if configmanager.GetInstance().Mongo.Atlas {
		connectionURI := fmt.Sprintf(connectionStringAtlas, configmanager.GetInstance().Mongo.Username, configmanager.GetInstance().Mongo.Password, configmanager.GetInstance().Mongo.Host, configmanager.GetInstance().Mongo.DBName, configmanager.GetInstance().Mongo.AppName)
		client, err = mongo.NewClient(baseOpts.ApplyURI(connectionURI))

		if err != nil {
			return nil, errors.New("failed to create client")
		}

	} else if configmanager.GetInstance().Mongo.SecureMongo {
		connectionURI := fmt.Sprintf(connectionStringMain, configmanager.GetInstance().Mongo.Username, configmanager.GetInstance().Mongo.Password, configmanager.GetInstance().Mongo.Host+":"+configmanager.GetInstance().Mongo.Port, configmanager.GetInstance().Mongo.DBName, readPreference)
		tlsConfig, err := u.getCustomTLSConfig(configmanager.GetInstance().Mongo.CertFile)
		if err != nil {
			return nil, errors.New("unable to get tls config")
		}
		client, err = mongo.NewClient(baseOpts.ApplyURI(connectionURI).SetTLSConfig(tlsConfig))
		if err != nil {
			return nil, errors.New("failed to create client")
		}
	} else {
		connectionURI := fmt.Sprintf(connectionStringDev, configmanager.GetInstance().Mongo.Host+":"+configmanager.GetInstance().Mongo.Port, configmanager.GetInstance().Mongo.DBName)
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
	u.dbName = configmanager.GetInstance().Mongo.DBName
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
	config := configmanager.GetInstance().Mongo
	return ConnectionInfo{
		DatabaseType: basetypes.MONGO,
		Host:         config.Host,
		Port:         config.Port,
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
