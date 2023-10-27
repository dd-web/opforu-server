package types

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
