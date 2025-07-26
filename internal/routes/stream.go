package routes

import (
	"fmt"
	"net/http"

	"EverythingSuckz/fsb/internal/config"
	"github.com/gin-gonic/gin"
)

// Stream renderiza el reproductor HTML con el enlace del archivo directo
func Stream(c *gin.Context) {
	fileID := c.Param("fileID")
	fileName := c.Param("fileName")

	if fileID == "" || fileName == "" {
		c.String(http.StatusBadRequest, "Parámetros inválidos.")
		return
	}

	// Genera la URL del archivo que será usada por el reproductor
	// Esta URL debe ser manejada por un handler que sirva el archivo (como /dl/:fileID/:fileName)
	baseURL := config.Get().BaseURL
	streamURL := fmt.Sprintf("%s/dl/%s/%s", baseURL, fileID, fileName)

	// Renderiza el template del reproductor
	c.HTML(http.StatusOK, "stream.html", gin.H{
		"StreamURL": streamURL,
		"FileName":  fileName,
	})
}
