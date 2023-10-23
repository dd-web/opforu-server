package handlers

import (
	"net/http"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/utils"
	"github.com/gorilla/mux"
)

/***********************************************************************************************/
/* ROOT path: host.com/api/thread/{slug}
/***********************************************************************************************/
func (rh *RoutingHandler) RegisterThreadRoot(w http.ResponseWriter, r *http.Request) error {
	queryCfg := utils.NewQueryConfig(r, "threads")

	switch r.Method {
	case "GET":
		return rh.handleThreadRoot(w, r, queryCfg)
	default:
		return HandleUnsupportedMethod(w, r)
	}
}

// GET: host.com/api/thread/{slug}
func (rh *RoutingHandler) handleThreadRoot(w http.ResponseWriter, r *http.Request, q *utils.QueryConfig) error {
	vars := mux.Vars(r)
	pipeline := builder.QrStrEntireThread(vars["slug"], q)

	thread, err := rh.Store.RunAggregation("threads", pipeline)
	if err != nil {
		return err
	}

	return HandleSendJSON(w, http.StatusOK, thread)
}
