package types

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/dd-web/opforu-server/internal/utils"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	MAX_FILE_SIZE_IMAGE = 8 * 1024 * 1024  // 8MB
	MAX_FILE_SIZE_VIDEO = 24 * 1024 * 1024 // 24MB

	FILE_NAME_CHAR_SET = "abcdefghijkmnpqrstuvwxyz123456789-_"
	FILE_NAME_LENGTH   = 32
)

// available asset types that can be uploaded - open for expansion
type AssetType string

const (
	AssetTypeImage AssetType = "image"
	AssetTypeVideo AssetType = "video"
)

func (at AssetType) String() string {
	return string(at)
}

type HashMethod string

const (
	HashMethodMD5    HashMethod = "md5"
	HashMethodSHA256 HashMethod = "sha256"
)

func (hm HashMethod) String() string {
	return string(hm)
}

type AssetSourceDetails struct {
	Avatar *FileCtx `json:"avatar" bson:"avatar"`
	Source *FileCtx `json:"source" bson:"source"`
}

// the internal source asset. This is a source file from which all other assets are derived. This should NEVER be passed
// to the client unless they are an admin. This is for privacy and storage reasons. The derived asset should be
// populated with the information the client needs from this source asset.
type AssetSource struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`

	Details *AssetSourceDetails `bson:"details" json:"details"`

	AssetType AssetType            `json:"asset_type" bson:"asset_type"`
	Uploaders []primitive.ObjectID `json:"uploaders" bson:"uploaders"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// the client level asset struct - assets are aggregated from asset sources with all the information they need.
// if a user attempts to upload a file with an existing hash, instead it will not be uploaded and instead just an
// Asset reference will be created. This is to prevent duplicate files from being uploaded and to let them name
// their file whatever they want for their own organization.
type Asset struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`

	SourceID  primitive.ObjectID `json:"source_id" bson:"source_id"`
	AccountID primitive.ObjectID `json:"account_id" bson:"account_id"`

	Description string   `json:"description" bson:"description"`
	FileName    string   `json:"file_name" bson:"file_name"`
	Tags        []string `json:"tags" bson:"tags"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

func NewSourceAsset() *AssetSource {
	ts := time.Now().UTC()
	as := &AssetSource{
		ID: primitive.NewObjectID(),
		Details: &AssetSourceDetails{
			Avatar: NewFileCtx(),
			Source: NewFileCtx(),
		},
		Uploaders: []primitive.ObjectID{},
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}

	return as
}

// creates a new asset, must provide an asset source id from which to derive and the account id of who uploaded it (or owns it)
func NewAsset(src primitive.ObjectID, acct primitive.ObjectID) *Asset {
	ts := time.Now().UTC()
	return &Asset{
		ID:          primitive.NewObjectID(),
		AccountID:   acct,
		SourceID:    src,
		Tags:        []string{},
		FileName:    "",
		Description: "",
		CreatedAt:   &ts,
		UpdatedAt:   &ts,
	}
}

// clone an existing asset but replace the account_id with a new one
func CloneAsset(asset *Asset, account_id primitive.ObjectID) *Asset {
	newAsset := NewAsset(asset.SourceID, account_id)
	newAsset.Description = asset.Description
	newAsset.FileName = asset.FileName
	newAsset.Tags = asset.Tags
	return newAsset
}

// Details about the file, since files can have avatar and source files this is abstracted out. Both may not need all of these.
type FileCtx struct {
	ServerFileName string `json:"server_file_name" bson:"server_file_name"`
	Height         uint16 `json:"height" bson:"height"`
	Width          uint16 `json:"width" bson:"width"`
	FileSize       uint32 `json:"file_size" bson:"file_size"`
	URL            string `json:"url" bson:"url"`
	Extension      string `json:"extension" bson:"extension"`
	HashMD5        string `json:"hash_md5" bson:"hash_md5"`
	HashSHA256     string `json:"hash_sha256" bson:"hash_sha256"`
}

// a new file context
func NewFileCtx() *FileCtx {
	return &FileCtx{
		ServerFileName: "",
		Height:         0,
		Width:          0,
		FileSize:       0,
		URL:            "",
		Extension:      "",
		HashMD5:        "",
		HashSHA256:     "",
	}
}

// returns an md5 checksum of the file located at filename
func GetFileChecksumMD5(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}

// returns an sha256 checksum of the file located at filename
func GetFileChecksumSHA256(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(hash.Sum(nil)), nil
}

// returns the size of the file located at filename
func GetFileSize(filename string) (int64, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}

	return file.Size(), nil
}

type FileUploadDetails struct {
	AssetType AssetType
	LocalID   string
	FileSize  uint32
	Height    int
	Width     int
}

func ParseFormFileDetails(rq *http.Request) *FileUploadDetails {
	details := &FileUploadDetails{
		AssetType: AssetType(rq.FormValue("type")),
		LocalID:   rq.FormValue("local_id"),
		Height:    0,
		Width:     0,
		FileSize:  0,
	}

	if height, err := strconv.Atoi(rq.FormValue("height")); err == nil {
		details.Height = height
	}

	if width, err := strconv.Atoi(rq.FormValue("width")); err == nil {
		details.Width = width
	}

	return details
}

func NewFileName() (string, error) {
	return gonanoid.Generate(FILE_NAME_CHAR_SET, FILE_NAME_LENGTH)
}

// upload temp file to spaces and return the url
func UploadFileToSpaces(tasset *utils.TempAsset) (string, error) {
	tempFile, err := os.Open(tasset.Dir)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	sess := session.Must(session.NewSession(utils.NewS3Config()))
	uploader := s3manager.NewUploader(sess)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String("opforu"),
		Key:    aws.String(fmt.Sprintf("%ss/%d%s", tasset.AssetType, tasset.TimeStamp, tasset.Ext)),
		Body:   tempFile,
		ACL:    aws.String("public-read"),
	})

	if err != nil {
		return "", err
	}

	return result.Location, nil
}
