package types

import (
	"crypto/md5"
	"crypto/sha256"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	MAX_FILE_SIZE_IMAGE = 8 * 1024 * 1024  // 8MB
	MAX_FILE_SIZE_VIDEO = 24 * 1024 * 1024 // 24MB
)

// available asset types that can be uploaded - open for expansion
type AssetType string

const (
	AssetTypeImage AssetType = "image"
	AssetTypeVideo AssetType = "video"
)

// the internal source asset. This is a source file from which all other assets are derived. This should NEVER be passed
// to the client unless they are an admin. This is for privacy and storage reasons. The derived asset should be
// populated with the information the client needs from this source asset.
type AssetSource struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`

	Details struct {
		Avatar FileCtx `json:"avatar" bson:"avatar"`
		Source FileCtx `json:"source" bson:"source"`
	}

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

// creates a new asset, must provide an asset source id from which to derive and the account id of who uploaded it (or owns it)
func NewAsset(src primitive.ObjectID, acct primitive.ObjectID) *Asset {
	ts := time.Now().UTC()
	return &Asset{
		ID:        primitive.NewObjectID(),
		AccountID: acct,
		SourceID:  src,
		Tags:      []string{},
		FileName:  "",
		CreatedAt: &ts,
		UpdatedAt: &ts,
	}
}

// Details about the file, since files can have avatar and source files this is abstracted out. Both may not need all of these.
type FileCtx struct {
	Height    uint64 `json:"height" bson:"height"`
	Width     uint64 `json:"width" bson:"width"`
	FileSize  uint64 `json:"file_size" bson:"file_size"`
	URL       string `json:"url" bson:"url"`
	Extension string `json:"extension" bson:"extension"`
}

// a new file context
func NewFileCtx() *FileCtx {
	return &FileCtx{
		Height:    0,
		Width:     0,
		FileSize:  0,
		URL:       "",
		Extension: "",
	}
}

// returns an md5 checksum of the file located at filename
func GetFileChecksumMD5(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := md5.New()

	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil

}

// returns an sha256 checksum of the file located at filename
func GetFileChecksumSHA256(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
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
	AssetType   AssetType
	LocalID     string
	Description string
	FileName    string
	Height      int
	Width       int
}

func ParseFormFileDetails(rq *http.Request) *FileUploadDetails {
	details := &FileUploadDetails{
		AssetType:   AssetType(rq.FormValue("type")),
		LocalID:     rq.FormValue("local_id"),
		Description: rq.FormValue("description"),
		FileName:    rq.FormValue("name"),
		Height:      0,
		Width:       0,
	}

	if height, err := strconv.Atoi(rq.FormValue("height")); err == nil {
		details.Height = height
	}

	if width, err := strconv.Atoi(rq.FormValue("width")); err == nil {
		details.Width = width
	}

	return details
}
