package types

import (
	"crypto/md5"
	"crypto/sha256"
	"io"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssetChecksum struct {
	MD5    []byte `bson:"md5" json:"md5"`
	SHA256 []byte `bson:"sha256" json:"sha256"`
}

type AssetAvatar struct {
	ID primitive.ObjectID `bson:"_id" json:"_id"`

	URL           string    `bson:"url" json:"url"`
	FileExtension string    `bson:"file_extension" json:"file_extension"`
	FileSize      int64     `bson:"file_size" json:"file_size"`
	AssetType     AssetType `bson:"asset_type" json:"asset_type"`

	Checksum AssetChecksum `bson:"checksum,omitempty" json:"checksum,omitempty"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type AssetSource struct {
	ID primitive.ObjectID `bson:"_id" json:"_id"`

	URL           string `bson:"url" json:"url"`
	FileExtension string `bson:"file_extension" json:"file_extension"`
	FileSize      int64  `bson:"file_size" json:"file_size"`

	AssetType AssetType   `bson:"asset_type" json:"asset_type"`
	Avatar    AssetAvatar `bson:"avatar,omitempty" json:"avatar,omitempty"`

	UploaderIDs []primitive.ObjectID `bson:"uploaded_by" json:"uploaded_by"`

	Checksum AssetChecksum `bson:"checksum" json:"checksum"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

// this is what gets sent along in requests - if there is a duplicate AssetSource hash then we'll make a new
// asset for the user and link it to the existing AssetSource
type Asset struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Source primitive.ObjectID `bson:"source" json:"source"`

	UserFileName string      `bson:"user_file_name" json:"user_file_name"`
	FileSize     int         `bson:"file_size" json:"file_size"`
	URL          int         `bson:"url" json:"url"`
	AssetType    AssetType   `bson:"asset_type" json:"asset_type"`
	Avatar       AssetAvatar `bson:"avatar" json:"avatar"`

	Checksum AssetChecksum `bson:"checksum" json:"checksum"`

	CreatedAt *time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt *time.Time `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

type AssetType string // an attempt to future proof if more asset types arise

const (
	AssetTypeImage AssetType = "image"
	AssetTypeVideo AssetType = "video"
)

func NewAsset() *Asset {
	ts := time.Now().UTC()
	return &Asset{
		ID:        primitive.NewObjectID(),
		CreatedAt: &ts,
		UpdatedAt: &ts,
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
