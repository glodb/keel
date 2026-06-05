package utils

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	mrand "math/rand"

	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/glodb/keel/httpHandler/basemodels"
	"github.com/glodb/keel/httpHandler/controllers/baseinterfaces"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/customtypes"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utilsdatatypes"

	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var getInstance = sync.OnceValue(func() *Utils {
	instance := &Utils{}
	instance.rng = mrand.New(mrand.NewSource(time.Now().UnixNano()))
	instance.digits = map[int]string{
		0: "0", 1: "1", 2: "2", 3: "3", 4: "4", 5: "5", 6: "6", 7: "7", 8: "8", 9: "9",
	}
	return instance
})

type Utils struct {
	digits map[int]string
	rng    *mrand.Rand
}

// GetInstance returns the singleton Utils instance.
func GetInstance() *Utils {
	return getInstance()
}

func (u *Utils) CreateMigrations(controller baseinterfaces.Controller, object basemodels.MigrationModels) []mongo.WriteModel {

	filePath := configmanager.GetInstance().MigrationsPath + "/" + string(controller.GetCollectionName()) + ".json"
	filePath = strings.ToLower(filePath)
	file, err := os.Open(filePath)
	if err != nil {
		logger.Log().Error("Error opening file", logger.StringField("filePath", filePath), logger.ErrorField("error", err))
		return []mongo.WriteModel{}
	}
	defer file.Close()

	jsonArray := make([]map[string]interface{}, 0)
	err = json.NewDecoder(file).Decode(&jsonArray)
	if err != nil {
		logger.Log().Error("Error decoding JSON from file", logger.StringField("filePath", filePath), logger.ErrorField("error", err))
		return []mongo.WriteModel{}
	}

	mongoQueue := make([]mongo.WriteModel, 0)
	for _, jsonData := range jsonArray {
		object.CleanUp()
		object.MapData(jsonData)
		operation := mongo.NewUpdateOneModel()
		operation.SetFilter(object.GetQuery())

		operation.SetUpdate(map[string]interface{}{"$set": object.GetUpdate()})
		operation.SetUpsert(true)
		mongoQueue = append(mongoQueue, operation)
	}
	return mongoQueue
}

func (u *Utils) GenerateUUID() (string, error) {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		return "", err
	}
	// Set the version (4) and variant (RFC4122) bits
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf(
			"%x-%x-%x-%x-%x",
			uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]),
		nil
}

// GenerateRandomNumber generates a random number with the specified number of digits
func (u *Utils) GenerateRandomNumber(digits int) (string, error) {

	randomNumber := ""
	for i := 0; i < digits; i++ {
		n := u.rng.Intn(10) // Generate a random number between 0 and 9
		randomNumber += u.digits[n]
	}

	return randomNumber, nil
}

// GenerateXID generates a unique XID (Extended ID) based on the current timestamp and a random number.
func (u *Utils) GenerateXID() string {
	guid := xid.New()
	return guid.String()
}

func (u *Utils) CopyMap(m map[string]*utilsdatatypes.Queue) map[string][]interface{} {
	cp := make(map[string][]interface{})
	for k, v := range m {
		cp[k] = v.Copy()
	}

	return cp
}

func (u *Utils) ToBsonDict(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func (u *Utils) NewUpdateOneModel(filter, update customtypes.M, insert bool) mongo.WriteModel {
	return mongo.NewUpdateOneModel().
		SetFilter(filter).
		SetUpdate(update).
		SetUpsert(insert)
}

func (u *Utils) Encrypt(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	// Create a new IV (Initialization Vector)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// CBC mode of operation
	mode := cipher.NewCBCEncrypter(block, iv)

	// Pad the plaintext to match block size
	paddedText := u.pad([]byte(plainText), aes.BlockSize)

	// Encrypt the plaintext
	ciphertext := make([]byte, len(paddedText))
	mode.CryptBlocks(ciphertext, paddedText)

	// Append IV to the ciphertext (needed for decryption)
	ciphertext = append(iv, ciphertext...)

	// Return as a base64 encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Padding for AES (PKCS7)
func (u *Utils) pad(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

func (u *Utils) Oauth2HTTPClient(ctx context.Context, creds *google.Credentials) *http.Client {
	ts := creds.TokenSource
	return oauth2.NewClient(ctx, ts)
}

func (u *Utils) ReadKeyFile(filePath string) []byte {
	data, err := os.ReadFile(filePath)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file: %v", err))
	}
	return data
}
