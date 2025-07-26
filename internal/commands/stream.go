package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"net/url"
	"strings"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
)

func (m *command) LoadStream(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("stream")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewMessage(streamHandler))
}

func streamHandler(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)

	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	message := u.EffectiveMessage()
	if message.Media == nil {
		ctx.Reply(u, "Please send a media file to get a streaming link.", nil)
		return dispatcher.EndGroups
	}

	fileId := fmt.Sprintf("%d", message.ID)
	streamURL := fmt.Sprintf("%s/%s", strings.TrimRight(config.ValueOf.BaseURL, "/"), url.PathEscape(fileId))

	replyText := fmt.Sprintf("🔗 **Here is your streaming link:**\n\n👉 %s", streamURL)
	ctx.Reply(u, replyText, nil)

	return dispatcher.EndGroups
}
