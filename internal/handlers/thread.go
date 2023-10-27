package handlers

import (
	"net/http"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
)

/***********************************************************************************************/
/* ROOT path: host.com/api/thread/{slug}
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterThreadRoot(rc *types.RequestCtx) error {
	// queryCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return rh.handleThreadRoot(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/thread/{slug}
func (rh *RoutingHandler) handleThreadRoot(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)
	pipeline := builder.QrStrEntireThread(vars["slug"], rc.Query)

	thread, err := rh.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return err
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, thread)
}
