package handlers

import (
	"net/http"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
)

type ThreadHandler struct {
	rc *types.RequestCtx
	rh *types.RoutingHandler
}

func InitThreadHandler(rh *types.RoutingHandler) *ThreadHandler {
	return &ThreadHandler{
		rh: rh,
	}
}

func (th *ThreadHandler) UpdateCtx(rc *types.RequestCtx) {
	th.rc = rc
}

/***********************************************************************************************/
/* ROOT path: host.com/api/thread/{slug}
/***********************************************************************************************/
func (th *ThreadHandler) RegisterThreadRoot(rc *types.RequestCtx) error {
	th.UpdateCtx(rc)
	// queryCfg := utils.NewQueryConfig(r, "threads")

	switch rc.Request.Method {
	case "GET":
		return th.handleThreadRoot(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/thread/{slug}
func (th *ThreadHandler) handleThreadRoot(rc *types.RequestCtx) error {
	th.UpdateCtx(rc)
	vars := mux.Vars(rc.Request)
	pipeline := builder.QrStrEntireThread(vars["slug"], rc.Query)

	thread, err := th.rh.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return err
	}

	return HandleSendJSON(rc.Writer, http.StatusOK, thread)
}
