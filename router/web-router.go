package router

import (
	"bytes"
	"embed"
	"encoding/base64"
	"html"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// WebAssets holds the embedded dashboard frontend assets.
type WebAssets struct {
	BuildFS   embed.FS
	IndexPage []byte
}

func SetWebRouter(router *gin.Engine, assets WebAssets, pluginDispatcher gin.HandlerFunc) {
	frontendFS := common.EmbedFolder(assets.BuildFS, "web/dist")
	compress := gzip.Gzip(gzip.DefaultCompression)

	// huiliu: the site icon follows the Logo option (data URI or URL); the
	// bundled file is only the fallback when no logo is configured.
	siteIcon := func(c *gin.Context) {
		logo := common.Logo
		if rest, ok := strings.CutPrefix(logo, "data:"); ok {
			meta, payload, found := strings.Cut(rest, ",")
			mimeType, isBase64 := strings.CutSuffix(meta, ";base64")
			if found && isBase64 && strings.HasPrefix(mimeType, "image/") {
				if decoded, err := base64.StdEncoding.DecodeString(payload); err == nil {
					c.Header("Cache-Control", "public, max-age=3600")
					c.Header("X-Content-Type-Options", "nosniff")
					c.Header("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'")
					c.Data(http.StatusOK, mimeType, decoded)
					return
				}
			}
		}
		if strings.HasPrefix(logo, "https://") || strings.HasPrefix(logo, "http://") {
			c.Redirect(http.StatusFound, logo)
			return
		}
		c.FileFromFS(c.Request.URL.Path, frontendFS)
	}
	router.GET("/favicon.ico", siteIcon)
	router.GET("/logo.png", siteIcon)

	router.NoRoute(
		pluginDispatcher,
		middleware.RouteTag("web"),
		func(c *gin.Context) {
			// ServeFrontendFiles sends build files already compressed.
			if !frontendFS.Exists("/", c.Request.URL.Path) {
				compress(c)
			}
		},
		middleware.AccessTokenAudit(),
		middleware.GlobalWebRateLimit(frontendFS),
		middleware.Cache(),
		middleware.ServeFrontendFiles(frontendFS),
		func(c *gin.Context) {
			if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
				controller.RelayNotFound(c)
				return
			}
			c.Header("Cache-Control", "no-cache")
			// huiliu: the page title comes from the SystemName option.
			siteName := []byte(html.EscapeString(common.SystemName))
			c.Data(http.StatusOK, "text/html; charset=utf-8", bytes.ReplaceAll(assets.IndexPage, []byte("__SITE_NAME__"), siteName))
		},
	)
}
