package cmd

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/routes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer()
	},
}

func startServer() {
	// Cargar configuración
	port := config.GetPort()

	// Iniciar router de Gin
	r := gin.Default()

	// Cargar templates
	r.LoadHTMLGlob("templates/*.html")

	// Rutas existentes
	r.GET("/", routes.Home)
	r.GET("/file/:fileID", routes.Download)

	// ✅ Nueva ruta para streaming con reproductor
	r.GET("/stream/:fileID/:fileName", routes.Stream)

	// Iniciar servidor con contexto para apagado limpio
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	// Canal para recibir señales del sistema
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ejecutar servidor en goroutine
	go func() {
		log.Printf("Servidor escuchando en http://localhost:%d", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error al iniciar el servidor: %s\n", err)
		}
	}()

	// Esperar señal para apagado
	<-quit
	log.Println("Apagando servidor...")

	// Contexto para timeout de apagado
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Error en apagado del servidor: %s\n", err)
	}
	log.Println("Servidor apagado correctamente")
}
