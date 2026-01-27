package swagger

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func SwaggerSetup() {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := exec.LookPath("swag"); err != nil {
			fmt.Printf("[swagger] swag CLI not found.  Install with: go install github.com/swaggo/swag/cmd/swag@latest")
		} else {
			var out bytes.Buffer
			cmd := exec.Command("swag", "init", "-g", "main.go", "-o", "docs")
			cmd.Dir = cwd
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Run(); err != nil {
				fmt.Printf("[swagger] swag init failed: %v\n%s", err, out.String())
			}
		}
	}
}

func GetDocPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "docs", "swagger.json")
}
