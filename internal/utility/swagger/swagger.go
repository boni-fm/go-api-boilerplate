package swagger

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SwaggerSetup jalanin swag init buat generate/update docs swagger
// kalau swag CLI gak ketemu, kasih tau cara installnya
func SwaggerSetup() {
	cwd, err := os.Getwd()
	if err == nil {
		if _, err := exec.LookPath("swag"); err != nil {
			fmt.Printf("[swagger] swag CLI gak ketemu. Install dulu: go install github.com/swaggo/swag/cmd/swag@latest")
		} else {
			var out bytes.Buffer
			cmd := exec.Command("swag", "init", "-g", "main.go", "-o", "docs")
			cmd.Dir = cwd
			cmd.Stdout = &out
			cmd.Stderr = &out
			if err := cmd.Run(); err != nil {
				fmt.Printf("[swagger] swag init gagal: %v\n%s", err, out.String())
			}
		}
	}
}

// GetDocPath kembaliin path lengkap ke file swagger.json
func GetDocPath() string {
	cwd, _ := os.Getwd()
	return filepath.Join(cwd, "docs", "swagger.json")
}
