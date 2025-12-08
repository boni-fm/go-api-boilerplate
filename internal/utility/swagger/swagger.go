package swagger

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func SwaggerSetup(app *fiber.App) {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := exec.LookPath("swag"); err != nil {
			// log.Printf("[swagger] swag CLI not found.  Install with: go install github.com/swaggo/swag/cmd/swag@latest")
		} else {
			var out bytes.Buffer
			cmd := exec.Command("swag", "init", "-g", "main. go", "-o", "docs")
			cmd.Dir = cwd
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Run(); err != nil {
				// log.Printf("[swagger] swag init failed: %v\n%s", err, out.String())
			} else {
				// log.Printf("[swagger] docs regenerated.\n%s", out. String())
			}
		}
	} else {
		// log.Printf("[swagger] failed to get working dir: %v", err)
	}

	// doc := redoc.Redoc{
	// 	Title:       "go-api-boilerplate API Documentation",
	// 	Description: "This is the API documentation for the go-api-boilerplate project.",
	// 	SpecFile:    filepath.Join(cwd, "docs", "swagger.json"),
	// 	SpecPath:    "/swagger.json",
	// 	DocsPath:    "/swagger",
	// }

	// app.Use(fiberredoc.New(doc))

	// Serve the generated JSON file directly
	docJSON := filepath.Join(cwd, "docs", "swagger.json")
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.SendFile(docJSON, true)
	})

	// Serve the Swagger UI at /swagger/*
	app.Get("/swagger/*", swagger.New(swagger.Config{
		URL:          "/swagger/doc.json",
		DeepLinking:  true,
		DocExpansion: "none",
	}))
}
