package handlers

import (
	"fmt"

	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
)

type ArticleHandler struct {
	rh *types.RoutingHandler
}

func InitArticleHandler(rh *types.RoutingHandler) *ArticleHandler {
	return &ArticleHandler{
		rh: rh,
	}
}

/***********************************************************************************************/
/* ROOT path: host.com/api/article
/***********************************************************************************************/
func (ah *ArticleHandler) RegisterArticleRoot(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return ah.handleArticleList(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// GET: host.com/api/article
func (ah *ArticleHandler) handleArticleList(rc *types.RequestCtx) error {
	articlepipe := builder.QrStrLookupArticle(rc.Query)

	count, err := rc.Store.CountArticleMatch(rc.Query.Search)
	if err != nil {
		fmt.Println("Error counting articles", err)
		return err
	}

	rc.Pagination.Update(int(count))

	articles, err := rc.Store.RunAggregation("articles", articlepipe)
	if err != nil {
		return err
	}

	rc.Records = articles
	// return HandleSendJSON(rc.Writer, http.StatusOK, bson.M{"articles": articles, "paginator": rc.Pagination})
	// rc.AddToResponseList("articles", articles)
	return ResolveResponse(rc)
}
