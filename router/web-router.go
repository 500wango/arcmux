package router

import (
	"embed"
	"net/http"
	"strings"

	"github.com/500wango/arcmux/common"
	"github.com/500wango/arcmux/controller"
	"github.com/500wango/arcmux/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

// WebAssets holds the embedded dashboard frontend assets.
type WebAssets struct {
	BuildFS   embed.FS
	IndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets WebAssets) {
	frontendFS := common.EmbedFolder(assets.BuildFS, "web/dist")

	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", frontendFS))
	router.NoRoute(func(c *gin.Context) {
		c.Set(middleware.RouteTagKey, "web")
		requestPath := c.Request.URL.Path
		if isMissingFrontendAsset(requestPath) {
			c.Header("Cache-Control", "no-store")
			c.Status(http.StatusNotFound)
			return
		}
		if strings.HasPrefix(requestPath, "/v1") || strings.HasPrefix(requestPath, "/api") || strings.HasPrefix(requestPath, "/assets") {
			controller.RelayNotFound(c)
			return
		}
		// The HTML shell contains content-hashed asset URLs. Do not cache it,
		// otherwise a deployment can leave the browser pointing at removed chunks.
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", assets.IndexPage)
	})
}

func isMissingFrontendAsset(path string) bool {
	switch path {
	case "/favicon.ico",
		"/logo.png",
		"/logo.svg",
		"/logo-icon.svg",
		"/logo-full.svg":
		return true
	}
	return strings.HasPrefix(path, "/static/")
}
