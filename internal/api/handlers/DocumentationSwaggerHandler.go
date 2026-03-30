package handlers

import (
	"go-api-boilerplate/internal/utility/swagger"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// GetDocumentation godoc
// @Summary Swagger API Documentation
// @Description Returns the Swagger/OpenAPI documentation in JSON format with proxy path support
// @Tags Swagger
// @Produce json
// @Success 200 {object} map[string]interface{} "Swagger documentation"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /swagger/doc.json [get]
func (h *HandlersRegistry) GetSwaggerDocumentation(c fiber.Ctx) error {
	// Get proxy path from context (set by middleware)
	proxyPath := swagger.GetProxyPath(c)

	// Get modified swagger document
	doc, err := h.SwaggerDoc.GetModifiedDocument(proxyPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load swagger documentation",
		})
	}

	return c.JSON(doc)
}

// GetUI godoc
// @Summary Swagger UI
// @Description Interactive Swagger UI for API documentation
// @Tags Swagger
// @Produce html
// @Success 200 {string} string "Swagger UI HTML"
// @Router /swagger [get]
func (h *HandlersRegistry) GetSwaggerUI(c fiber.Ctx) error {
	// Get proxy path from context
	proxyPath := swagger.GetProxyPath(c)

	// Build doc URL with proxy path
	docURL := "/swagger/doc.json"
	if proxyPath != "" {
		proxyPath = strings.TrimSuffix(proxyPath, "/")
		docURL = proxyPath + "/swagger/doc.json"
	}

	// httpSwagger.Handler returns a net/http handler. In Fiber v3, the built-in
	// adaptor middleware bridges net/http handlers into Fiber's handler chain
	// without requiring manual fasthttp-to-net/http conversion.
	return adaptor.HTTPHandler(httpSwagger.Handler(
		httpSwagger.URL(docURL),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
	))(c)
}
