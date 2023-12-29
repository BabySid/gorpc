package http

import (
	"github.com/BabySid/gorpc/api"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
)

func getHandleWrapper(handle api.RawHandle) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		id := uuid.New().String()
		if v, ok := ctx.GetQuery("id"); ok {
			id = v
		}
		myCtx := newRawContext(path, id, 0, ctx)
		defer func() {
			myCtx.EndRequest(api.Success)
		}()

		handle(myCtx, nil)
	}
}

func postHandleWrapper(handle api.RawHandle) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path

		id := uuid.New().String()
		if v, ok := ctx.GetQuery("id"); ok {
			id = v
		}
		myCtx := newRawContext(path, id, 0, ctx)
		defer func() {
			myCtx.EndRequest(api.Success)
		}()

		body, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.String(http.StatusBadRequest, "read body err: %v", err)
			return
		}

		handle(myCtx, body)
	}
}
