package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SwaggerHandler 处理 Swagger 文档相关请求
type SwaggerHandler struct {
	docPath string
}

// NewSwaggerHandler 创建新的 Swagger 处理器
func NewSwaggerHandler(docPath string) *SwaggerHandler {
	return &SwaggerHandler{
		docPath: docPath,
	}
}

// GetSwaggerJSON 返回 Swagger JSON 文档
func (h *SwaggerHandler) GetSwaggerJSON(c *gin.Context) {
	swaggerFile := filepath.Join(h.docPath, "swagger.json")
	c.File(swaggerFile)
}

// GetSwaggerUI 返回 Swagger UI 页面
func (h *SwaggerHandler) GetSwaggerUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Laojun API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/swagger/doc.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}