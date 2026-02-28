package core

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"main.go/global"
	"main.go/initialize"
	"main.go/router"
)

func RunWebhookServer() {
	// Initialize K8s client and informer
	initialize.K8s()

	r := gin.New()
	r.Use(gin.Recovery())

	// Initialize webhook router
	router.RouterGroupApp.Webhook.InitWebhookRouter(r.Group("/"))

	host := global.GVA_CONFIG.System.Host
	port := global.GVA_CONFIG.System.WebhookPort
	address := fmt.Sprintf("%s:%d", host, port)

	log.Printf("Starting Webhook server on %s...", address)

	certFile := global.GVA_CONFIG.System.TlsCert
	keyFile := global.GVA_CONFIG.System.TlsKey

	// Fallback to default paths if not configured
	if certFile == "" {
		certFile = "/etc/webhook/certs/tls.crt"
	}
	if keyFile == "" {
		keyFile = "/etc/webhook/certs/tls.key"
	}

	// Certificates paths from config
	if err := r.RunTLS(address, certFile, keyFile); err != nil {
		log.Fatalf("Webhook server failed: %v", err)
	}
}
