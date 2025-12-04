package main

import (
	"fmt"

	"github.com/boni-fm/go-libsd3/pkg/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/mikhail-bigun/fiberlogrus"
)

const (
	appName = "api-boilerplate"
	isDebug = true
)

func main() {

	// init log
	log_ := log.NewLoggerWithFilename(appName)
	log_.Logger.SetFormatter(log_.Formatter)

	app := fiber.New(fiber.Config{
		AppName: appName,
	})

	// ✨ Middleware
	// > Logrus logger middleware
	app.Use(
		fiberlogrus.New(
			fiberlogrus.Config{
				Logger: log_.Logger,
				Tags:   fiberlogrus.CommonTags,
			},
		))

	// > Recover middleware
	app.Use(
		recover.New(
			recover.Config{
				EnableStackTrace: true,
			},
		))

	// > Healthcheck middleware
	app.Use(
		healthcheck.New(
			healthcheck.Config{
				LivenessProbe: func(c *fiber.Ctx) bool {
					return true
				},
				LivenessEndpoint: "/live",
			},
		),
	)

	// > 404 not found handler
	app.Static("/", "./static/public")
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).SendFile("./static/public/404.html")
	})

	// Mulai ~ 🤩

	fmt.Println("Service started ~~ ༼ つ ◕_◕ ༽つ")
	fmt.Println(`
		                   ▒░                        ▒▒             
		                    ▒░░░░                ▒░░░▒▒             
		                    ▒▒▒░░░░░▒░░░░░░░░░▒░░░░▒▒▒              
		                     ▒▒▒░░░░░░░░░░░░░░░░░░▒▒▒▒              
		                      ░░░░░░░░░░░░░░░░░░░░░░▒░░░░▒          
		                     ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░       
		                  ░░░░░░░▒▒▒░░░░░░░░░░░▒▒▒░░░░▒░░░░░░░░░    
		               ░░░░░░░░░░░░░░░░▒▓▓▒░░░░▒▓▓▒░░░░░░░░░░░░░░░  
		▒           ░░░░░░▒░░▒░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ 
		 ░▒    ▒░░░░░░░░░░░░░▒░░░░░░░░░░░░░░░░░░░░░░▒░░░░░░░░░░░░░░░
		   ▒▒▒░░░░░▒▒▒▒▒▒▒░░▒▒▒░░░░░░░░░░░░░░░░░░░░▒░░░░░░░░░░░░░░░▒
		        ▒▒▒▒▒▒▒▒▒░░░▒▒▒▒▒▒░░░░░░░░░░░░░░░▒░░░▒░░░░░░░░░░░░░ 
		            ▒    ░░░▒        ▒▒▒▒▒▒▒▒ ▒▒░░░░░▒░░░░░░▒▒░░░░  
		             ░  ▒░░░▒ ▒      ░▒▒▒▒▒▒▒░ ░░░░▒▒░▒░░░░░▒▒░░░   
		               ▒░▒░▒   ▒    ░░░░░░░░░░░░░░▒░░░▒▒░▒▒░░░░▒    
		               ░░░▒         ░░░░░░░░░░░░░░░░▒▒░▒▒░▓▒▒▒      
		              ▒▒░░▒        ▒░░░░░░░░░░░░░░░░░░░░▒▓▒▒▒       
		             ▒▒▒▓▒         ░░░░░░░░░░░░░░▒▒▒▒▒▒▒▒▒▒▒▒▒▒     
		           ▒▒▒▒▒▒         ░░▒░░░▒▒░░░░░░▒▒▒▒▒      ▒▒▒▒▒    
		          ▒▒▒▒▒▒          ░░▒░░░▒▒░░░░░░░           ▒▒▒▒    
		          ▒▒▒▒▒          ▒░░░░░░▒▒░░░▒░░░                   
		           ▒▒            ▒░░░░░░▒▒░░░░░░▒                   
		                          ▒░░▒░░▒▒░░▒░░░▒                   
		                           ▒▒▒▒▒▒▒▒▒▒▒▒▒                    
		                            ▒▒▒▒▒▒▒▒▒▒                      
		                               ▒▓▒▒▒                        
	`)

	app.Listen(":8080")
}
