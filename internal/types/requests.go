package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type RequestUnmarshaller interface {
	UnmarshalFromReqInto(*RequestCtx) error
}

/***********************************************************************************************/
/* Request UnMarshaller structs - used to unmarshal request bodies into objects
/***********************************************************************************************/

// Session will be an ID in the cookie
type RUMSession struct {
	SessionID string `json:"session"`
}

type RUMAssetAttachment struct {
	SourceID    primitive.ObjectID `json:"source_id"`
	Description string             `json:"description,omitempty"`
	FileName    string             `json:"file_name,omitempty"`
	Tags        []string           `json:"tags,omitempty"`
}

func NewRUMAssetAttachment() *RUMAssetAttachment {
	return &RUMAssetAttachment{
		SourceID:    primitive.ObjectID{},
		Description: "",
		FileName:    "",
		Tags:        []string{},
	}
}

// new thread requests come through the board/[short] POST route we already have the board because
// it's in the endpoint. these are the rest of the fields on a request to create a new thread
type RUMThread struct {
	Title   string               `json:"title"`
	Content string               `json:"content"`
	Assets  []RUMAssetAttachment `json:"assets"`
}

func NewRUMThread() *RUMThread {
	return &RUMThread{
		Title:   "",
		Content: "",
		Assets:  make([]RUMAssetAttachment, 0),
	}
}
