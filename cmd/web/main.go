package main

import (
	"flag"
	"flagged-it/internal/api"
	"log"
	"net/http"
)

func main() {
	port := flag.String("port", "8080", "Port to serve on")
	dev := flag.Bool("dev", false, "Run in development mode")
	flag.Parse()

	// Setup API routes
	api.SetupRoutes()

	// Serve assets directory (flags, icons, etc.)
	assetsPath := "../../assets"
	fs := http.FileServer(http.Dir(assetsPath))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	if *dev {
		log.Println("Serving assets from filesystem:", assetsPath)
		log.Println("Development mode: Svelte dev server should be running on port 5173")
	} else {
		log.Println("Production mode: Build frontend first with 'make web-build'")
		// TODO: Serve embedded frontend in production
	}

	log.Printf("API server starting on http://0.0.0.0:%s (accessible from network)\n", *port)
	log.Printf("   Local: http://localhost:%s\n", *port)
	log.Println("API endpoints: /api/*")
	log.Println("Assets: /assets/*")
	log.Println()

	if err := http.ListenAndServe("0.0.0.0:"+*port, nil); err != nil {
		log.Fatal(err)
	}
}
