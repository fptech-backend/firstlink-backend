package utils

import (
	"archive/tar"
	"bytes"
	"certification/config"
	"certification/constant"
	"certification/logger"
	"certification/model"
	"context"
	"reflect"
	"strconv"

	"compress/gzip"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/otiai10/copy"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	age "github.com/theTardigrade/golang-age"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type RedisValue struct {
	Token  string
	Module []map[string]interface{}
	Status string
}

type Initializer struct {
	DB          *gorm.DB
	RDB         *redis.Client
	MDB         *mongo.Database
	FB          *firestore.Client
	S3          *s3.Client
	DaysInYear  Year
	DaysInMonth []Month
	Location    time.Location
}

type Year struct {
	Year int
	Days float32
}

type Month struct {
	Month int
	Days  int
}

var fileTypes = map[string]string{
	"data:image/png;base64,":     "png",
	"data:image/jpg;base64,":     "jpg",
	"data:image/jpeg;base64,":    "jpeg",
	"data:image/bmp;base64,":     "bmp",
	"data:image/svg+xml;base64,": "svg",
	"data:image/x-icon;base64,":  "ico",
}

func IsValidUUID(u string, id *uuid.UUID) bool {
	var err error
	*id, err = uuid.Parse(u)
	return err == nil
}

func GenerateToken() (string, error) {
	// Generate a 16-byte random token
	token := make([]byte, 16)

	_, err := rand.Read(token)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %v", err)
	}

	// Encode the token in base64 format
	encodedToken := base64.URLEncoding.EncodeToString(token)

	return encodedToken, nil
}

func GenerateJWT(id *uuid.UUID, profile_id *uuid.UUID, email *string, expiry *time.Time, role *constant.AccountRoleType) (string, error) {
	tokenJWT := jwt.New(jwt.SigningMethodHS256)
	claims := tokenJWT.Claims.(jwt.MapClaims)
	claims["id"] = *id
	claims["profile_id"] = *profile_id
	claims["email"] = *email
	claims["exp"] = expiry.Unix()
	claims["role"] = *role

	// Generate tokenJWT (JWT) with a secret key
	token, err := tokenJWT.SignedString([]byte(config.SECRET))

	return token, err
}

// Validate token from token table whether it's exist and return account id
func ValidateToken(token string, db *gorm.DB) (uuid.UUID, error) {
	var tokenData model.Token
	err := db.Where("token = ?", token).First(&tokenData).Error
	if err != nil {
		return uuid.Nil, err
	}

	return tokenData.AccountID, nil
}

// Update the token status to used
func UpdateTokenStatus(token string, tx *gorm.DB) error {
	err := tx.Model(&model.Token{}).Where("token = ?", token).Update("status", constant.USED).Error
	if err != nil {
		return err
	}

	return nil
}

// Generate random 6-digit OTP
func GenerateOTPToken(accountID uuid.UUID, tx *gorm.DB) (string, error) {
	otp := make([]byte, 6)
	_, err := rand.Read(otp)

	if err != nil {
		return "", fmt.Errorf("failed to generate OTP token: %v", err)
	}

	for i := 0; i < 6; i++ {
		otp[i] = uint8(48 + (otp[i] % 10))
	}

	logger.Log.Info(time.Now(), " : ", time.Now().Add(time.Minute*5))

	//store otp in token
	token := model.Token{
		AccountID: accountID,
		Token:     string(otp),
		ExpireAt:  time.Now().Add(time.Minute * 5), // 5 minutes expiry
		Type:      constant.OTP_TOKEN,
		Status:    constant.PENDING,
	}

	err = tx.Create(&token).Error
	if err != nil {
		tx.Rollback()
		logger.Log.Error(err)
		return "", fmt.Errorf("failed to store OTP token: %v", err)
	}

	return string(otp), nil
}

// --------------- DB Table ---------------

func IsTableEmpty(db *gorm.DB, table string) bool {
	var count int
	query := fmt.Sprintf("SELECT COUNT(ID) FROM %s", table)
	db.Raw(query).Scan(&count)

	return count == 0
}

func IsValueExisting(db *gorm.DB, table string, key string, value interface{}) bool {
	var count int
	query := fmt.Sprintf("SELECT COUNT(ID) FROM %s WHERE %s = ?", table, key)
	db.Raw(query, value).Scan(&count)

	return count == 0
}

// --------------- Redis ---------------

func (initializer *Initializer) UpdateStatusInRedis(results string, id string) error {
	var data RedisValue
	err := json.Unmarshal([]byte(results), &data)
	if err != nil {
		return fmt.Errorf("error in decoding string to json for ID %s. %s", id, results)
	}

	data.Status = constant.UPDATED
	jsonData, err := jsoniter.Marshal(data)
	if err != nil {
		return fmt.Errorf("error in encoding JSON %s for ID %s", err, id)
	}

	// Store in Redis
	err = initializer.RDB.Set(context.Background(), id, jsonData, 0).Err()
	if err != nil {
		return fmt.Errorf("unable to set EmpID %s and JSON %v to Redis", id, data)
	}

	return nil
}

func (initializer *Initializer) UpdateObjectHSetInRedis(key string, value interface{}) error {
	return initializer.RDB.HSet(context.Background(), key, "data", value).Err()
}

func (initializer *Initializer) GetObjectHGetAllByKeyInRedis(key string) string {
	return initializer.RDB.HGet(context.Background(), key, "data").Val()
}

// --------------- Web 3 ---------------
func VerifyWallet() error {
	pvKeyHex := ""

	pvKeyBytes, err := hex.DecodeString(pvKeyHex)
	if err != nil {
		return err
	}

	pvKey, err := crypto.ToECDSA(pvKeyBytes)
	if err != nil {
		return err
	}

	message := []byte("hello world")
	r, s, err := ecdsa.Sign(rand.Reader, pvKey, message)
	if err != nil {
		return err
	}

	isVerified := ecdsa.Verify(&pvKey.PublicKey, message, r, s)
	if !isVerified {
		return errors.New("Invalid wallet info")
	}

	return nil
}

// --------------- Date Time ---------------

func ParseToTime(timeString string) (time.Time, error) {
	convertedTime, err := time.ParseInLocation(constant.TIME_FORMAT_24, timeString, time.Local)
	if err != nil {
		convertedTime, err = time.ParseInLocation(constant.TIME_FORMAT_12, timeString, time.Local)
	}

	return convertedTime, err
}

func GetTotalDaysOfCurrentMonth(month int) int {
	today := time.Now()

	firstDayOfMonth := time.Date(today.Year(), time.Month(month), 1, 0, 0, 0, 0, today.Location())
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)

	return lastDayOfMonth.Day()
}

func GetTotalDaysOfCurrentYear() float32 {
	today := time.Now()

	firstDayOfYear := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, today.Location())
	firstDayOfNextYear := time.Date(today.Year()+1, time.January, 1, 0, 0, 0, 0, today.Location())

	return float32(firstDayOfNextYear.Sub(firstDayOfYear).Hours() / 24)
}

func (initializer *Initializer) UpdateTotalDaysOfCurrentYear() {
	initializer.DaysInYear.Year = time.Now().Year()
	initializer.DaysInYear.Days = GetTotalDaysOfCurrentYear()

	initializer.DaysInMonth = nil

	for i := 1; i <= 12; i++ {
		initializer.DaysInMonth = append(initializer.DaysInMonth, Month{
			Month: i,
			Days:  GetTotalDaysOfCurrentMonth(i),
		})
	}
}

func CalculateYearsToNow(date time.Time) float32 {
	return float32(age.CalculateToNow(date))
}

// --------------- File ---------------

func GetImageFileType(base64String string) (string, string, error) {
	for key, value := range fileTypes {
		if strings.HasPrefix(base64String, key) {
			format := value
			if value == "svg" {
				format += "+xml"
			} else if value == "ico" {
				format = "x-icon"
			}
			return value, format, nil
		}
	}

	return "", "", errors.New("unsupported image format")
}

func LocateFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return err
	}

	return nil
}

func UploadImage(prevImagePath string, base64String string, imageName string, childPath string) (string, error) {
	if prevImagePath != "" {
		if err := LocateFile(prevImagePath); err != nil {
			return "", err
		}
	}

	ext, format, err := GetImageFileType(base64String)
	if err != nil {
		return "", err
	}

	rawImageString := strings.TrimPrefix(base64String, "data:image/"+format+";base64,")

	decodedImage, err := base64.StdEncoding.DecodeString(rawImageString)
	if err != nil {
		return ext, err
	}

	// create folder if not exist
	if err := LocateFile("image"); err != nil {
		if err := os.Mkdir("image", os.ModePerm); err != nil {
			return "", err
		}
	}

	if err := LocateFile("image/" + childPath); err != nil {
		if err := os.Mkdir("image/"+childPath, os.ModePerm); err != nil {
			return "", err
		}
	}

	imagePath := "image/" + childPath + "/" + imageName + "." + ext
	file, err := os.Create(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(decodedImage)
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

func DownloadImage(imagePath string) string {
	if err := LocateFile(imagePath); err != nil {
		logger.Log.Info(err)
		return ""
	}

	var headerBase64 string
	for key, value := range fileTypes {
		if filepath.Ext(imagePath)[1:] == value {
			headerBase64 = key
			break
		}
	}

	imageByte, err := os.ReadFile(imagePath)
	if err != nil {
		logger.Log.Info(err)
		return ""
	}
	return headerBase64 + base64.StdEncoding.EncodeToString(imageByte)
}

func DeleteFile(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}

	return nil
}

func GetAllFilePaths(dir string) ([]string, error) {
	var filePaths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			filePaths = append(filePaths, filepath.Base(path))
		}

		return nil
	})

	return filePaths, err
}

func CompressAndEncodeFiles(dir string, filePaths []string) (string, int, error) {
	// Compress the files into a gzip-compressed archive
	var compressedData bytes.Buffer
	var base64String string
	gzipWritter := gzip.NewWriter(&compressedData)
	tarWritter := tar.NewWriter(gzipWritter)

	for _, filePath := range filePaths {
		finalFilePath := filepath.Join(dir, filePath)

		file, err := os.Open(finalFilePath)
		if err != nil {
			return base64String, fiber.StatusBadRequest, fmt.Errorf("error opening file: %s", err)
		}
		defer file.Close()

		// Prepare the file info for the archive
		fileInfo, _ := file.Stat()
		header := &tar.Header{
			Name:    filePath,
			Size:    fileInfo.Size(),
			Mode:    int64(fileInfo.Mode()),
			ModTime: fileInfo.ModTime(),
		}

		// Write the file info and content to the archive
		if err := tarWritter.WriteHeader(header); err != nil {
			return base64String,
				fiber.StatusInternalServerError,
				fmt.Errorf("error writing tar header: %s", err)
		}
		if _, err := io.Copy(tarWritter, file); err != nil {
			return base64String,
				fiber.StatusInternalServerError,
				fmt.Errorf("error writing file to archive: %s", err)
		}
	}

	// Close the archive and gzip writer
	if err := tarWritter.Close(); err != nil {
		return base64String,
			fiber.StatusInternalServerError,
			fmt.Errorf("error closing tar writer: %s", err)
	}

	if err := gzipWritter.Close(); err != nil {
		return base64String,
			fiber.StatusInternalServerError,
			fmt.Errorf("error closing gzip writer: %s", err)
	}

	// Encode the compressed data into base64
	base64String = base64.StdEncoding.EncodeToString(compressedData.Bytes())

	return base64String, fiber.StatusOK, nil
}

func DecodeAndSaveCompressedFiles(
	base64String string,
	saveDir string,
	supportFileExt *map[string]bool,
	state ...string,
) (int, error) {
	isUpdateState := len(state) > 0 && state[0] == "update"

	decodedData, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		logger.Log.Error(err)
		return fiber.StatusBadRequest, fmt.Errorf("base64 decoding error")
	}

	// Create gzip reader
	gzipReader, err := gzip.NewReader(bytes.NewReader(decodedData))
	if err != nil {
		logger.Log.Error(err)
		return fiber.StatusBadRequest, fmt.Errorf("error creating gzip reader")
	}

	defer gzipReader.Close()

	// Create tar reader to read individual files
	tarReader := tar.NewReader(gzipReader)

	// Directory to save the decompressed files
	err = LocateFile(saveDir)
	isNewDir := err != nil
	backupDir := saveDir + "_backup"

	if isUpdateState && !isNewDir {
		err := copy.Copy(saveDir, backupDir)
		if err != nil {
			errMsg := "error create backup files"
			logger.Log.Error(errMsg, err)
			return fiber.StatusInternalServerError, fmt.Errorf(errMsg)
		}

		DeleteFile(saveDir)
	}

	if err = os.MkdirAll(saveDir, os.ModePerm); err != nil {
		logger.Log.Error(err, saveDir)
		return fiber.StatusInternalServerError, fmt.Errorf("error creating save directory")
	}

	errMsg := ""

	for {
		tarHeader, err := tarReader.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			errMsg = "error reading tar header"
			logger.Log.Error(err)
			break
		}

		// Create a new file to save the decompressed content
		savePath := filepath.Join(saveDir, tarHeader.Name)

		saveExt := filepath.Ext(savePath)[1:]

		_, exist := (*supportFileExt)[saveExt]
		if !exist {
			errMsg = fmt.Sprintf("unsupported file format: %s", saveExt)
			logger.Log.Error(errMsg, savePath)
			break
		}

		saveFile, err := os.Create(savePath)
		if err != nil {
			errMsg = "error creating save file"
			logger.Log.Error(err, savePath)
			saveFile.Close()
			break
		}

		// Write the decompressed content to the save file
		if _, err := io.Copy(saveFile, tarReader); err != nil {
			errMsg = "error writing save file"
			logger.Log.Error(err, savePath)
			saveFile.Close()
			break
		}

		saveFile.Close()

		logger.Log.Infof("Uploaded file is saved:", savePath)
	}

	if errMsg != "" {
		DeleteFile(saveDir)

		if isUpdateState && !isNewDir {
			err := os.Rename(backupDir, saveDir)

			if err != nil {
				logger.Log.Error(err, backupDir)
			}
		}

		return fiber.StatusInternalServerError, fmt.Errorf(errMsg)
	}

	if isUpdateState && !isNewDir {
		DeleteFile(backupDir)
	}

	return fiber.StatusOK, nil
}

// --------------- Round ---------------

func RoundToOneDecimalPlaces(number float64) float64 {
	numberStr := strconv.FormatFloat(number*10, 'f', 1, 32)
	numberFloat64, _ := strconv.ParseFloat(numberStr, 32)
	return math.Round(numberFloat64) / 10
}

func RoundToTwoDecimalPlaces(number float64) float64 {
	return math.Round(number*100) / 100
}

func RoundSecondDecimalValueTo5(value float64) float64 {
	tempValue := int(value * 100)
	secondDecimalValue := tempValue % 10

	if secondDecimalValue != 0 {
		value = float64(tempValue-secondDecimalValue+5) / 100
	}

	return value
}

func RoundUpToTheNearest0_05(value float64) float64 {
	return math.Round(value*20+0.5) / 20
}

// --------------- Range ---------------

func GetAverageInRange(min float32, max float32) float64 {
	average := (max + min) / 2
	return float64(average)
}

func GenerateRangeWithCommonDifference(
	startRange float64,
	endRange float64,
	commonDifference float64,
	arrRange *[]float64) {
	for i := startRange; i < endRange; i += commonDifference {
		*arrRange = append(*arrRange, i)
	}
}

func GenerateMinAndMaxRange(
	ranges *map[string][]float32,
	start float32,
	end float32,
	gap float32,
) {
	if *ranges == nil {
		*ranges = make(map[string][]float32)
	}

	increment := start
	for i := 0; increment < end; i++ {
		(*ranges)["min"] = append((*ranges)["min"], increment+0.01)
		increment += gap
		(*ranges)["max"] = append((*ranges)["max"], increment)

		average := float32(GetAverageInRange(
			(*ranges)["min"][i],
			(*ranges)["max"][i],
		))

		(*ranges)["average"] = append((*ranges)["average"], average)
	}
}

// --------------- Validator ---------------
func ValidateParser(body interface{}, ctx *fiber.Ctx, tagName string) error {
	if err := ctx.BodyParser(&body); err != nil {
		return err
	}

	validate := validator.New()
	validate.SetTagName(tagName)

	if err := validate.Struct(body); err != nil {
		return err
	}

	return nil
}

func ValidateFormParser(form interface{}, ctx *fiber.Ctx) error {
	if err := ctx.BodyParser(form); err != nil {
		return err
	}

	validate := validator.New()

	if err := validate.Struct(form); err != nil {
		return err
	}

	return nil
}

func UpdateBSONParser(val *reflect.Value, bodyMap *map[string]interface{}, validateTag string) (primitive.M, error) {
	var (
		validate = validator.New()
		typ      = val.Type()
		update   = bson.M{"$set": bson.M{}}
	)

	for i := 0; i < val.NumField(); i++ {
		var (
			field = val.Field(i)
			value = field.Interface() // variable's value

			fieldType   = typ.Field(i)
			fieldName   = fieldType.Tag.Get("json")      // json tag
			validateTag = fieldType.Tag.Get(validateTag) // validate tag
		)

		// Verify existence of field
		if _, ok := (*bodyMap)[fieldName]; !ok {
			continue
		}

		// Variable validation
		if validateTag != "" && validateTag != "-" {
			if err := validate.Var(value, validateTag); err != nil {
				return primitive.M{}, fmt.Errorf("validation error for field [%s]: %s", fieldName, err)
			}
		}

		// append the verified field into BSON
		update[constant.MONGO_SET].(bson.M)[fmt.Sprintf(constant.MONGO_FIELDS+".$.%s", fieldName)] = value
	}

	return update, nil
}

// hexToRGB takes a hex color code as a string and returns the RGB representation.
func HexToRGB(hexColor string) (uint8, uint8, uint8, error) {
	// Remove the hash (#) character if it's present
	if hexColor[0] == '#' {
		hexColor = hexColor[1:]
	}

	// Hex color code should be 6 characters
	if len(hexColor) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color length")
	}

	// Parse the hex color string to integers
	r, err := strconv.ParseInt(hexColor[0:2], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	g, err := strconv.ParseInt(hexColor[2:4], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}
	b, err := strconv.ParseInt(hexColor[4:6], 16, 64)
	if err != nil {
		return 0, 0, 0, err
	}

	return uint8(r), uint8(g), uint8(b), nil
}
