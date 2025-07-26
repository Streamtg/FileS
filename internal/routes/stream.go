package routes

import (
	"fmt"
	"net/http"

	"EverythingSuckz/fsb/config"
	"github.com/gin-gonic/gin"
)

func Stream(c *gin.Context) {
	fileID := c.Param("fileID")
	fileName := c.Param("fileName")

	if fileID == "" || fileName == "" {
		c.String(http.StatusBadRequest, "Faltan parámetros en la URL.")
		return
	}

	streamURL := fmt.Sprintf("%s/dl/%s/%s", config.GetBaseURL(), fileID, fileName)

	c.HTML(http.StatusOK, "stream.html", gin.H{
		"fileID":    fileID,
		"fileName":  fileName,
		"streamURL": streamURL,
	})
}
