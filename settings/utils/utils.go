package utils

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
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

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
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

// Singleton. Returns a single object of Utils
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

// generateRequestID creates a fast random request ID
func (u *Utils) GenerateRequestID() string {
	bytes := make([]byte, 8) // 16 hex characters
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
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

func (u *Utils) IsTokenValid(tokenString string) (bool, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Provide the key for token validation
		return []byte(configmanager.GetInstance().SessionSecret), nil
	})

	if err != nil {
		return false, err
	}

	// Check if the token is valid
	if token.Valid {
		// Token is valid, check the expiration time
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
			currentTime := time.Now()

			if expirationTime.After(currentTime) {
				return true, nil
			} else {
				return false, errors.New("token has expired")
			}
		} else {
			return false, errors.New("error extracting claims")
		}
	} else {
		return false, errors.New("error parsing token")
	}
}

func (u *Utils) ToBsonDict(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

// GenerateHashedPassword generates a hashed password from the given plaintext password.
func (u *Utils) GenerateHashedPassword(password string) (string, error) {
	// Generate the hashed password using bcrypt with default cost.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func (u *Utils) GetStartOfDay(t int64) int64 {
	t = t - (t % (24 * 60 * 60))
	return t
}

func (u *Utils) GetSNearestThirtyMinutes(t int64) int64 {

	minutes := t % (30 * 60)
	if minutes > 0 {
		t = t - minutes
	}

	return t
}

func (u *Utils) GetStartOfMonth(t time.Time) int64 {
	year, month, _ := t.Date()
	location := t.Location()
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, location)
	return startOfMonth.Unix()
}

func (u *Utils) RemoveAsterik(arr []string) []string {
	index := -1
	for i, val := range arr {
		if val == "*" {
			index = i
			break
		}
	}

	// If "*" was found, remove it from the array
	if index != -1 {
		arr = append(arr[:index], arr[index+1:]...)
	}
	return arr
}

func (u *Utils) RemoveStringFromArray(arr []string, str string) []string {
	index := -1
	for i, val := range arr {
		if val == str {
			index = i
			break
		}
	}

	// If string was found, remove it from the array
	if index != -1 {
		arr = append(arr[:index], arr[index+1:]...)
	}
	return arr
}

func (u *Utils) NewUpdateOneModel(filter, update customtypes.M, insert bool) mongo.WriteModel {
	return mongo.NewUpdateOneModel().
		SetFilter(filter).
		SetUpdate(update).
		SetUpsert(insert)
}

func (u *Utils) ValidateMeasurementTime(timestamps ...int64) bool {
	currentTime := time.Now().UTC().Unix()

	currentTime += 12 * 60 * 60 //In order to support GMT, we allow to add readings for GMT+12

	// Helper function to check if a timestamp is valid
	isValid := func(ts int64) bool {
		logger.Log().Debug("ValidateMeasurementTime", logger.Int64Field("current_time", currentTime), logger.Int64Field("ts", ts))
		return ts != 0 && ts <= currentTime
	}
	// Map of validation functions for different cases
	validators := map[int]func([]int64) bool{
		2: func(ts []int64) bool {
			return isValid(ts[0]) && isValid(ts[1]) && ts[0] > ts[1]
		},
	}
	// Get the appropriate validator based on the number of timestamps
	if validator, exists := validators[len(timestamps)]; exists {
		return validator(timestamps)
	}
	// Invalid number of arguments
	return false
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
