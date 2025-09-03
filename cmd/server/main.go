// @title Server Management Service
// @version 1.0
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
package main

import (
	"context"

	"github.com/th1enq/ViettelSMS_ServerService/internal/application"
)

func main() {
	app, err := application.InitApp()
	if err != nil {
		panic("failed to initialize server: " + err.Error())
	}
	app.Start(context.Background())
}
