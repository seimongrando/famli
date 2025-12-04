package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"legacybridge/handlers"
	appmw "legacybridge/middleware"
	"legacybridge/store"
)

func main() {
	port := getenv("PORT", "8080")

	// Diretório do frontend (index.html, styles.css, app.js)
	// Ajuste se seu build estiver em outro caminho.
	staticDir := getenv("STATIC_DIR", filepath.Join("..", "frontend"))

	// Store em memória (MVP)
	memoryStore := store.NewMemoryStore()

	// Segredo JWT: compartilhado entre AuthHandler e middleware
	jwtSecret := os.Getenv("JWT_SECRET")

	authHandler := handlers.NewAuthHandler(memoryStore, jwtSecret)
	itemsHandler := handlers.NewItemsHandler(memoryStore)
	guardiansHandler := handlers.NewGuardiansHandler(memoryStore)
	settingsHandler := handlers.NewSettingsHandler(memoryStore)
	assistantHandler := handlers.NewAssistantHandler()

	// Router raiz
	r := chi.NewRouter()
	r.Use(
		chimiddleware.RequestID,
		chimiddleware.RealIP,
		chimiddleware.Logger,
		chimiddleware.Recoverer,
	)

	// API sob /api
	r.Route("/api", func(api chi.Router) {
		// Rotas públicas (sem auth)
		api.Post("/auth/register", authHandler.Register)
		api.Post("/auth/login", authHandler.Login)

		// Rotas protegidas (precisam de sessão JWT no cookie)
		api.Group(func(pr chi.Router) {
			pr.Use(appmw.NewJWTMiddleware(jwtSecret))

			pr.Get("/auth/me", authHandler.Me)
			pr.Post("/auth/logout", authHandler.Logout)

			pr.Get("/items", itemsHandler.List)
			pr.Post("/items", itemsHandler.Create)
			pr.Put("/items/{itemID}", itemsHandler.Update)
			pr.Delete("/items/{itemID}", itemsHandler.Delete)

			pr.Get("/guardians", guardiansHandler.List)
			pr.Post("/guardians", guardiansHandler.Create)
			pr.Delete("/guardians/{guardianID}", guardiansHandler.Delete)

			pr.Get("/settings", settingsHandler.Get)
			pr.Post("/settings", settingsHandler.Save)

			pr.Post("/assistant", assistantHandler.Handle)
		})
	})

	// Servir o SPA (frontend)
	fileServer := http.FileServer(http.Dir(staticDir))

	// Qualquer rota que não comece com /api cai no frontend
	r.Handle("/*", fileServer)

	log.Printf("Famli server escutando em http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
