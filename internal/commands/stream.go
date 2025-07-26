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
	"github.com/gotd/td/tg"
)

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("stream")
	defer log.Sugar().Info("Stream handler loaded")

	dispatcher.AddHandler(
		handlers.NewMessage(nil, sendLink),
	)
}

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

	if peer.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	if len(config.ValueOf.AllowedUsers) > 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	supported, err := supportedMediaFilter(u.EffectiveMessage)
	if err != nil {
		return err
	}
	if !supported {
		ctx.Reply(u, "Sorry, this message type is unsupported.", nil)
		return dispatcher.EndGroups
	}

	update, err := utils.ForwardMessages(ctx, chatId, config.ValueOf.LogChannelID, u.EffectiveMessage.ID)
	if err != nil {
		utils.Logger.Sugar().Errorf("ForwardMessages error: %v", err)
		ctx.Reply(u, fmt.Sprintf("Error forwarding message: %s", err.Error()), nil)
		return dispatcher.EndGroups
	}

	messageID := update.Updates[0].(*tg.UpdateMessageID).ID
	msgMedia := update.Updates[1].(*tg.UpdateNewChannelMessage).Message.(*tg.Message).Media

	file, err := utils.FileFromMedia(msgMedia)
	if err != nil {
		utils.Logger.Sugar().Errorf("FileFromMedia error: %v", err)
		ctx.Reply(u, fmt.Sprintf("Error extracting file info: %s", err.Error()), nil)
		return dispatcher.EndGroups
	}

	fullHash := utils.PackFile(file.FileName, file.FileSize, file.MimeType, file.ID)
	hash := utils.GetShortHash(fullHash)

	baseHost := config.ValueOf.Host
	if !strings.HasPrefix(baseHost, "http://") && !strings.HasPrefix(baseHost, "https://") {
		baseHost = "http://" + baseHost
	}

	link := fmt.Sprintf("%s/stream/%d?hash=%s", baseHost, messageID, hash)

	// Texto visible clickeable (sin estilo code para que Telegram lo reconozca)
	visibleText := fmt.Sprintf("Direct link: %s", link)

	// Construcción de botones con texto en mayúsculas
	row := tg.KeyboardButtonRow{
		Buttons: []tg.KeyboardButtonClass{
			&tg.KeyboardButtonURL{
				Text: "DOWNLOAD",
				URL:  link + "&d=true",
			},
		},
	}

	if strings.Contains(file.MimeType, "video") || strings.Contains(file.MimeType, "audio") || strings.Contains(file.MimeType, "pdf") {
		row.Buttons = append(row.Buttons, &tg.KeyboardButtonURL{
			Text: "STREAM",
			URL:  link,
		})
	}

	markup := &tg.ReplyInlineMarkup{Rows: []tg.KeyboardButtonRow{row}}

	// Enviar mensaje con texto visible y botones si no es localhost
	if strings.Contains(baseHost, "localhost") || strings.Contains(baseHost, "127.0.0.1") {
		_, err = ctx.Reply(u, []interface{}{visibleText}, &ext.ReplyOpts{
			NoWebpage:        false,
			ReplyToMessageId: u.EffectiveMessage.ID,
		})
	} else {
		_, err = ctx.Reply(u, []interface{}{visibleText}, &ext.ReplyOpts{
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
