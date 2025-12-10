package swagger

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/gofiber/fiber/v2"
)

func SwaggerSetup(app *fiber.App) {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := exec.LookPath("swag"); err != nil {
			// log.Printf("[swagger] swag CLI not found.  Install with: go install github.com/swaggo/swag/cmd/swag@latest")
		} else {
			var out bytes.Buffer
			cmd := exec.Command("swag", "init", "-g", "main.go", "-o", "docs")
			cmd.Dir = cwd
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Run(); err != nil {
				fmt.Printf("[swagger] swag init failed: %v\n%s", err, out.String())
			} else {
				fmt.Printf("[swagger] docs regenerated.\n%s", out.String())
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
}
