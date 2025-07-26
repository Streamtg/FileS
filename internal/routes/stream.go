package routes

import (
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gotd/td/tg"
	range_parser "github.com/quantumsheep/range-parser"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var log *zap.Logger

func (e *allRoutes) LoadHome(r *Route) {
	log = e.log.Named("Stream")
	defer log.Info("Loaded stream route")
	r.Engine.GET("/stream/:messageID", getStreamRoute)
}

func getStreamRoute(ctx *gin.Context) {
	w := ctx.Writer
	r := ctx.Request

	messageIDParm := ctx.Param("messageID")
	messageID, err := strconv.Atoi(messageIDParm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	authHash := ctx.Query("hash")
	if authHash == "" {
		http.Error(w, "missing hash param", http.StatusBadRequest)
		return
	}

	worker := bot.GetNextWorker()

	file, err := utils.FileFromMessage(ctx, worker.Client, messageID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expectedHash := utils.PackFile(
		file.FileName,
		file.FileSize,
		file.MimeType,
		file.ID,
	)
	if !utils.CheckHash(authHash, expectedHash) {
		http.Error(w, "invalid hash", http.StatusBadRequest)
		return
	}

	// Si es foto sin tamaño (thumbnail o foto)
	if file.FileSize == 0 {
		res, err := worker.Client.API().UploadGetFile(ctx, &tg.UploadGetFileRequest{
			Location: file.Location,
			Offset:   0,
			Limit:    1024 * 1024,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result, ok := res.(*tg.UploadFile)
		if !ok {
			http.Error(w, "unexpected response", http.StatusInternalServerError)
			return
		}
		fileBytes := result.GetBytes()
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.FileName))
		if r.Method != "HEAD" {
			ctx.Data(http.StatusOK, file.MimeType, fileBytes)
		}
		return
	}

	// Si el usuario solicita descarga forzada
	if ctx.Query("d") == "true" {
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.FileName))
	} else {
		// Si el tipo es video/audio/pdf, servir página con player
		if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") || strings.Contains(file.MimeType, "pdf") {
			streamURL := fmt.Sprintf("%s/stream/%d?hash=%s&d=true", utils.GetBaseURL(ctx), messageID, authHash)
			var tag string
			switch {
			case strings.Contains(file.MimeType, "video"):
				tag = "video"
			case strings.Contains(file.MimeType, "audio"):
				tag = "audio"
			case strings.Contains(file.MimeType, "pdf"):
				tag = "iframe"
			default:
				tag = "video"
			}

			html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta property="og:image" content="https://www.flaticon.com/premium-icon/icons/svg/2626/2626281.svg" itemprop="thumbnailUrl">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" type='text/css' href="https://drive.google.com/uc?export=view&id=1pVLG4gZy7jdow3sO-wFS06aP_A9QX0O6">
    <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Raleway">
    <link rel="stylesheet" href="https://fonts.googleapis.com/css?family=Delius">
    <link rel="stylesheet" href="https://cdn.plyr.io/3.5.6/plyr.css" />
</head>
<body>
    <header>
        <div class="toogle"></div>
        <div id="file-name" class="cyber">%s</div>
    </header>
    <div class="container">
        <%s src="%s" class="player" controls></%s>
    </div>
    <footer>
        <span id="Visit-text">Telegram Channel</span>
        <span>
            <a href="https://t.me/peligxg" id='Netflix-logo'>
                <svg id='octo' style="width: 1.2rem; padding-left: 5px; fill: var(--footer-icon-color)" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
            </a>
        </span>
    </footer>
    <script type="text/javascript">
	atOptions = {
		'key' : '72f2fe2e6c4a708c45e60e4c80b834c6',
		'format' : 'iframe',
		'height' : 250,
		'width' : 300,
		'params' : {}
	};
	document.write('<scr' + 'ipt type="text/javascript" src="http' + (location.protocol === 'https:' ? 's' : '') + '://allureencourage.com/72f2fe2e6c4a708c45e60e4c80b834c6/invoke.js"></scr' + 'ipt>');
</script>
<script async src="https://www.googletagmanager.com/gtag/js?id=G-XPCZ4QGYJT"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'G-XPCZ4QGYJT');
</script>
<script src="https://cdn.plyr.io/2.0.13/demo.js"></script>
<script src="https://cdn.plyr.io/3.5.6/plyr.js"></script>
<script>
    const controls = [
          'play-large', 'rewind', 'play', 'fast-forward', 'progress',
          'current-time', 'duration', 'mute', 'volume', 'captions',
          'settings', 'pip', 'airplay', 'fullscreen'
    ];
    document.addEventListener('DOMContentLoaded', () => {
        const player = Plyr.setup('.player', { controls });
    });
    const body = document.querySelector('body');
    const footer = document.querySelector('footer');
    const toogle = document.querySelector('.toogle');
    toogle.onclick = () => {
        body.classList.toggle('dark')
        footer.classList.toggle('dark')
    }
</script>
</body>
</html>`, file.FileName, file.FileName, tag, streamURL, tag)

			ctx.Header("Content-Type", "text/html; charset=utf-8")
			_, err = w.Write([]byte(html))
			if err != nil {
				log.Error("failed to write response", zap.Error(err))
			}
			return
		}
	}

	ctx.Header("Accept-Ranges", "bytes")
	var start, end int64
	rangeHeader := r.Header.Get("Range")

	if rangeHeader == "" {
		start = 0
		end = file.FileSize - 1
		w.WriteHeader(http.StatusOK)
	} else {
		ranges, err := range_parser.Parse(file.FileSize, r.Header.Get("Range"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		start = ranges[0].Start
		end = ranges[0].End
		ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, file.FileSize))
		log.Info("Content-Range", zap.Int64("start", start), zap.Int64("end", end), zap.Int64("fileSize", file.FileSize))
		w.WriteHeader(http.StatusPartialContent)
	}

	contentLength := end - start + 1
	mimeType := file.MimeType

	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	ctx.Header("Content-Type", mimeType)
	ctx.Header("Content-Length", strconv.FormatInt(contentLength, 10))

	// Ya seteamos Content-Disposition para descarga arriba si se pidió
	if ctx.Query("d") != "true" {
		ctx.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", file.FileName))
	}

	if r.Method != "HEAD" {
		lr, _ := utils.NewTelegramReader(ctx, worker.Client, file.Location, start, end, contentLength)
		if _, err := io.CopyN(w, lr, contentLength); err != nil {
			log.Error("Error while copying stream", zap.Error(err))
		}
	}
}
