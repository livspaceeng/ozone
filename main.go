package main

import (
	// "github.com/livspaceeng/ozone/configs"

	"github.com/livspaceeng/ozone/internal/server"
	log "github.com/sirupsen/logrus"
)

// @title           Ozone API
// @version         1.0
// @description     An auth layer for APIs
// @termsOfService  https://livspace.io

// @contact.name   Ankit
// @contact.url    https://livspace.io
// @contact.email  ankit.a@livspace.com

// @license.name  Apache 2.0
// @license.url   https://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1
// @schemes   http
func main() {
	// config.Init()
	log.SetFormatter(&log.JSONFormatter{})
	// log.SetOutput(os.Stdout)

	logLevel, err := log.ParseLevel(log.DebugLevel.String())
	if err != nil {
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
	server.Init()
}
