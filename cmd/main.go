package main

import (
	"log"

	"fixio/internal/app/server"

	_ "fixio/docs" // Swagger generated docs — registers spec
)

// @title           Fixio API — Forum Kritik & Solusi Kebijakan Publik
// @version         1.0
// @description     REST API untuk platform forum di mana masyarakat menyampaikan kritik terhadap kebijakan dan mengusulkan solusi.
// @host            localhost:8080
// @BasePath        /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	srv := server.NewServer()
	if err := srv.Start(); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
