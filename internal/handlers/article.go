package handlers

import (
	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"go.mongodb.org/mongo-driver/bson"
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
/* ROOT path: host.com/api/articles
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

// GET: host.com/api/articles
func (ah *ArticleHandler) handleArticleList(rc *types.RequestCtx) error {
	var count int64
	var pipeline bson.A
	var articles []bson.M

	pipeline = builder.QrStrLookupArticle(rc.Query)
	count = rc.Store.CountResults("articles", rc.Query.Search)

	articles, err := rc.Store.RunAggregation("articles", pipeline)
	if err != nil {
		return err
	}

	rc.Pagination.Update(int(count))
	rc.Records = articles

	return ResolveResponse(rc)
}
