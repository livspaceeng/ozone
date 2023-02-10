package main

import (
	// "github.com/livspaceeng/ozone/configs"

	"github.com/livspaceeng/ozone/internal/server"
	"github.com/livspaceeng/ozone/middleware"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/contrib/propagators/b3"
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
	traceProvider, err := middleware.JaegerTraceProvider()
	if err != nil {
		log.Error(err)
	}
	otel.SetTracerProvider(traceProvider)
	// b3propagator to track external server calls
	p := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader | b3.B3SingleHeader))
    otel.SetTextMapPropagator(p)
	server.Init()
}
