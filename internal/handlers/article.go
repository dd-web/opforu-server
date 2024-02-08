package handlers

import (
	"github.com/dd-web/opforu-server/internal/builder"
	"github.com/dd-web/opforu-server/internal/types"
	"github.com/gorilla/mux"
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

// METHOD: GET
// PATH: host.com/api/articles
func (ah *ArticleHandler) handleArticleList(rc *types.RequestCtx) error {
	var count int64
	var pipeline bson.A
	var articles []bson.M

	pipeline = builder.QrStrLookupArticleList(rc.Query)
	count = rc.Store.CountResults("articles", rc.Query.Search)

	articles, err := rc.Store.RunAggregation("articles", pipeline)
	if err != nil {
		return err
	}

	rc.Pagination.Update(int(count))
	rc.Records = articles

	return ResolveResponse(rc)
}

/***********************************************************************************************/
/* ROOT path: host.com/api/articles/{slug}
/***********************************************************************************************/
func (ah *ArticleHandler) RegisterArticleSlug(rc *types.RequestCtx) error {
	rc.UpdateStore(ah.rh.Store)

	switch rc.Request.Method {
	case "GET":
		return ah.handleArticleShort(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/articles/{slug}
func (ah *ArticleHandler) handleArticleShort(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)
	// articles := []bson.M{}

	pipeline := builder.QrStrLookupArticle(vars["slug"])

	article, err := rc.Store.RunAggregation("articles", pipeline)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("article", article[0])
	return ResolveResponse(rc)
}
