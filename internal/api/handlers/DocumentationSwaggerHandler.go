package handlers

import (
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

// GetSwaggerDocumentation godoc
// @Summary Swagger API Documentation
// @Description Returns the Swagger/OpenAPI documentation in JSON format
// @Tags _
// @Produce json
// @Success 200 {object} map[string]interface{} "Swagger documentation"
// @Router /swagger/doc.json [get]
func (hr *HandlersRegistry) GetSwaggerDocumentation(c *fiber.Ctx) error {
	cwd, _ := os.Getwd()
	return c.SendFile(filepath.Join(cwd, "docs", "swagger.json"), true)
}

// GetSwaggerUI godoc
// @Summary Swagger UI
// @Description Interactive Swagger UI for API documentation
// @Tags _
// @Produce html
// @Router /swagger [get]
func (hr *HandlersRegistry) GetSwaggerUI(c *fiber.Ctx) error {
	return swagger.New(
		swagger.Config{
			URL:          "/swagger/doc.json",
			DeepLinking:  true,
			DocExpansion: "none",
		},
	)(c)
}
