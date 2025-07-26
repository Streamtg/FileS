package commands

import (
	"fmt"
	"strings"

	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
	"github.com/celestix/gotgproto/types"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("stream")
	defer log.Sugar().Info("Stream handler loaded")

	// Agrega el manejador para mensajes entrantes
	dispatcher.AddHandler(
		handlers.NewMessage(nil, sendLink),
	)
}

// supportedMediaFilter revisa si el mensaje tiene un tipo de media soportado
func supportedMediaFilter(m *types.Message) (bool, error) {
	if m.Media == nil {
		return false, dispatcher.EndGroups
	}
	switch m.Media.(type) {
	case *tg.MessageMediaDocument, *tg.MessageMediaPhoto:
		return true, nil
	default:
		return false, nil
	}
}

func sendLink(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peer := ctx.PeerStorage.GetPeerById(chatId)

	// Solo usuarios directos pueden usar el bot
	if peer.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	// Control de acceso por lista blanca
	if len(config.ValueOf.AllowedUsers) > 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	// Verificar si el mensaje contiene media soportada
	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}
	if !supported {
		ctx.Reply(u, "Sorry, this message type is unsupported.", nil)
		return dispatcher.EndGroups
	}

	// Reenviar mensaje a canal log para obtener metadata
	update, err := utils.ForwardMessages(ctx, chatId, config.ValueOf.LogChannelID, u.EffectiveMessage.ID)
	if err != nil {
		utils.Logger.Sugar().Errorf("ForwardMessages error: %v", err)
		ctx.Reply(u, fmt.Sprintf("Error forwarding message: %s", err.Error()), nil)
		return dispatcher.EndGroups
	}

	// Obtener ID de mensaje reenviado y metadata del archivo
	messageID := update.Updates[0].(*tg.UpdateMessageID).ID
	msgMedia := update.Updates[1].(*tg.UpdateNewChannelMessage).Message.(*tg.Message).Media

	file, err := utils.FileFromMedia(msgMedia)
	if err != nil {
		utils.Logger.Sugar().Errorf("FileFromMedia error: %v", err)
		ctx.Reply(u, fmt.Sprintf("Error extracting file info: %s", err.Error()), nil)
		return dispatcher.EndGroups
	}

	// Generar hash corto para link seguro
	fullHash := utils.PackFile(file.FileName, file.FileSize, file.MimeType, file.ID)
	hash := utils.GetShortHash(fullHash)

	// Construir link con hash y puerto/host configurado
	baseHost := config.ValueOf.Host
	if !strings.HasPrefix(baseHost, "http://") && !strings.HasPrefix(baseHost, "https://") {
		baseHost = "http://" + baseHost
	}

	link := fmt.Sprintf("%s/stream/%d?hash=%s", baseHost, messageID, hash)

	// Construir texto con estilo monospace para link
	text := []styling.StyledTextOption{styling.Code(link)}

	// Crear botones inline para descarga y streaming
	row := tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "Download",
				URL:  link + "&d=true",
			},
		},
	}

	// Solo agregar botón streaming para ciertos tipos MIME
	if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") || strings.Contains(file.MimeType, "pdf") {
		row.Buttons = append(row.Buttons, &tg.KeyboardButtonURL{
			Text: "Stream",
			URL:  link,
		})
	}

	markup := &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{row}}

	// Evitar botones cuando se está en localhost para evitar confusión
	if strings.Contains(baseHost, "localhost") || strings.Contains(baseHost, "127.0.0.1") {
		_, err = ctx.Reply(u, text, &ext.ReplyOpts{
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	} else {
		_, err = ctx.Reply(u, text, &ext.ReplyOpts{
			Markup:           markup,
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	}

	if err != nil {
		utils.Logger.Sugar().Errorf("Reply error: %v", err)
		ctx.Reply(u, fmt.Sprintf("Error sending message: %s", err.Error()), nil)
	}

	return dispatcher.EndGroups
}
