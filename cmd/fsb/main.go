package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

// Esta estructura contiene los datos que enviaremos al template HTML
type FileTemplateData struct {
	FileName    string
	FileSize    string
	MimeType    string
	DownloadURL string
	IsVideo     bool
}

// Simulación de la estructura que devolvería la función que busca el archivo
type FileInfo struct {
	FileName string
	Size     int64
	MimeType string
}

// Esta función deberías reemplazarla con la real que obtenga el archivo desde Telegram
func GetFileInfoFromTelegram(fileID string) (*FileInfo, error) {
	// Aquí deberías usar tu lógica real de Pyrogram o Telethon o la API que tengas
	// Por ahora, devolvemos un archivo falso de prueba
	return &FileInfo{
		FileName: "ejemplo.mp4",
		Size:     10485760, // 10 MB
		MimeType: "video/mp4",
	}, nil
}

// Convierte bytes a formato legible (KB, MB, GB...)
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

// Handler que renderiza el HTML
func fileViewHandler(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Falta el parámetro 'id'", http.StatusBadRequest)
		return
	}

	fileInfo, err := GetFileInfoFromTelegram(fileID)
	if err != nil {
		http.Error(w, "Archivo no encontrado", http.StatusNotFound)
		return
	}

	tmpl, err := template.ParseFiles("templates/file.html")
	if err != nil {
		http.Error(w, "Error al cargar la plantilla", http.StatusInternalServerError)
		return
	}

	isVideo := strings.HasPrefix(fileInfo.MimeType, "video/")

	data := FileTemplateData{
		FileName:    fileInfo.FileName,
		FileSize:    ByteCountDecimal(fileInfo.Size),
		MimeType:    fileInfo.MimeType,
		DownloadURL: "/download?id=" + fileID,
		IsVideo:     isVideo,
	}

	tmpl.Execute(w, data)
}

// Simulación del handler de descarga
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "Falta el parámetro 'id'", http.StatusBadRequest)
		return
	}

	// Aquí deberías devolver el archivo real desde Telegram
	w.Header().Set("Content-Disposition", "attachment; filename=ejemplo.mp4")
	w.Header().Set("Content-Type", "video/mp4")
	w.Write([]byte("Este sería el contenido del archivo"))
}

func main() {
	http.HandleFunc("/view", fileViewHandler)
	http.HandleFunc("/download", downloadHandler)

	fmt.Println("Servidor iniciado en http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
