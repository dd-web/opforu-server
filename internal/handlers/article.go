package handlers

import (
	"encoding/json"
	"io"
	"time"

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
		return ah.handleGetSingleArticle(rc)
	case "POST":
		return ah.handleNewArticleComment(rc)
	default:
		return HandleUnsupportedMethod(rc.Writer, rc.Request)
	}
}

// METHOD: GET
// PATH: host.com/api/articles/{slug}
func (ah *ArticleHandler) handleGetSingleArticle(rc *types.RequestCtx) error {
	vars := mux.Vars(rc.Request)
	pipeline := builder.QrStrLookupArticle(vars["slug"])

	article, err := rc.Store.RunAggregation("articles", pipeline)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	if len(article) == 0 {
		return ResolveResponseErr(rc, types.ErrorNotFound("article not found"))
	}

	rc.AddToResponseList("article", article[0])
	return ResolveResponse(rc)
}

// METHOD: POST
// PATH: host.com/api/articles/{slug}
func (ah *ArticleHandler) handleNewArticleComment(rc *types.RequestCtx) error {
	if rc.UnresolvedAccount {
		return ResolveResponseErr(rc, types.ErrorUnauthorized())
	}

	// setup necessary dependencies
	vars := mux.Vars(rc.Request)
	ts := time.Now().UTC()
	details := types.NewRUMComment()

	article, err := rc.Store.FindArticleBySlug(vars["slug"])
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	body, err := io.ReadAll(rc.Request.Body)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = json.Unmarshal(body, &details)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	// dependency injection
	newCommentAssets := []*types.Asset{}
	newCommentAssetInterfaces := []interface{}{}

	for _, v := range details.Assets {
		a := types.NewAsset(v.SourceID, rc.AccountCtx.Account.ID)
		a.FileName = v.FileName
		a.Description = v.Description
		a.Tags = v.Tags
		newCommentAssets = append(newCommentAssets, a)
	}

	// @todo - replace with parsed body
	commentBody := details.Content

	comment := types.NewArticleComment()

	article.CommentRef++
	comment.CommentNumber = article.CommentRef
	comment.Body = commentBody
	comment.AuthorID = rc.AccountCtx.Account.ID

	comment.UpdatedAt = &ts
	article.UpdatedAt = &ts

	if details.MakeAnonymous && rc.AccountCtx.Account.IsStaff() {
		comment.AuthorAnon = true
	}

	for _, v := range newCommentAssets {
		comment.Assets = append(comment.Assets, v.ID)
		newCommentAssetInterfaces = append(newCommentAssetInterfaces, v)
	}

	article.Comments = append(article.Comments, comment.ID)

	// save & update associative docs
	err = rc.Store.SaveNewMulti(newCommentAssetInterfaces, "assets")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.SaveNewSingle(comment, "article_comments")
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	err = rc.Store.UpdateArticle(article)
	if err != nil {
		return ResolveResponseErr(rc, types.ErrorUnexpected())
	}

	rc.AddToResponseList("comment_number", comment.CommentNumber)
	return ResolveResponse(rc)
}
