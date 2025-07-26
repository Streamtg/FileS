package main

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/cache"
	"EverythingSuckz/fsb/internal/routes"
	"EverythingSuckz/fsb/internal/types"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use:                "run",
	Short:              "Run the bot with the given configuration.",
	DisableSuggestions: false,
	Run:                runApp,
}

var startTime time.Time = time.Now()

// Struct para pasar datos a la plantilla /view
type FileTemplateData struct {
	FileName    string
	FileSize    string
	MimeType    string
	DownloadURL string
	IsVideo     bool
}

// Función para formatear tamaño de archivo en forma legible
func ByteCountDecimal(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// Simulación de obtención de info de archivo (reemplazar por lógica real)
func GetFileInfoFromTelegram(fileID string) (*FileTemplateData, error) {
	// Ejemplo estático
	return &FileTemplateData{
		FileName:    "video-ejemplo.mp4",
		FileSize:    ByteCountDecimal(10485760), // 10 MB
		MimeType:    "video/mp4",
		DownloadURL: "/download?id=" + fileID,
		IsVideo:     strings.HasPrefix("video/mp4", "video/"),
	}, nil
}

func runApp(cmd *cobra.Command, args []string) {
	utils.InitLogger(config.ValueOf.Dev)
	log := utils.Logger
	mainLogger := log.Named("Main")
	mainLogger.Info("Starting server")
	config.Load(log, cmd)
	router := getRouter(log)

	mainBot, err := bot.StartClient(log)
	if err != nil {
		log.Panic("Failed to start main bot", zap.Error(err))
	}
	cache.InitCache(log)
	workers, err := bot.StartWorkers(log)
	if err != nil {
		log.Panic("Failed to start workers", zap.Error(err))
		return
	}
	workers.AddDefaultClient(mainBot, mainBot.Self)
	bot.StartUserBot(log)
	mainLogger.Info("Server started", zap.Int("port", config.ValueOf.Port))
	mainLogger.Info("File Stream Bot", zap.String("version", versionString))
	mainLogger.Sugar().Infof("Server is running at %s", config.ValueOf.Host)
	err = router.Run(fmt.Sprintf(":%d", config.ValueOf.Port))
	if err != nil {
		mainLogger.Sugar().Fatalln(err)
	}
}

func getRouter(log *zap.Logger) *gin.Engine {
	if config.ValueOf.Dev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.Use(gin.ErrorLogger())

	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, types.RootResponse{
			Message: "Server is running.",
			Ok:      true,
			Uptime:  utils.TimeFormat(uint64(time.Since(startTime).Seconds())),
			Version: versionString,
		})
	})

	// Nuevo endpoint /view para mostrar info del archivo con plantilla HTML
	router.GET("/view", func(ctx *gin.Context) {
		fileID := ctx.Query("id")
		if fileID == "" {
			ctx.String(http.StatusBadRequest, "Falta el parámetro 'id'")
			return
		}

		fileInfo, err := GetFileInfoFromTelegram(fileID)
		if err != nil {
			ctx.String(http.StatusNotFound, "Archivo no encontrado")
			return
		}

		tmpl, err := template.ParseFiles("templates/file.html")
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Error al cargar plantilla")
			return
		}

		ctx.Header("Content-Type", "text/html; charset=utf-8")
		err = tmpl.Execute(ctx.Writer, fileInfo)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Error al renderizar plantilla")
		}
	})

	routes.Load(log, router)
	return router
}
