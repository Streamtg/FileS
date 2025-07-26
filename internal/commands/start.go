package commands

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/utils"

	"github.com/celestix/gotgproto/dispatcher"
	"github.com/celestix/gotgproto/dispatcher/handlers"
	"github.com/celestix/gotgproto/ext"
	"github.com/celestix/gotgproto/storage"
)

func (m *command) LoadStart(dispatcher dispatcher.Dispatcher) {
	log := m.log.Named("start")
	defer log.Sugar().Info("Loaded")
	dispatcher.AddHandler(handlers.NewCommand("start", start))
}

func start(ctx *ext.Context, u *ext.Update) error {
	chatId := u.EffectiveChat().GetID()
	peerChatId := ctx.PeerStorage.GetPeerById(chatId)

	if peerChatId.Type != int(storage.TypeUser) {
		return dispatcher.EndGroups
	}

	if len(config.ValueOf.AllowedUsers) != 0 && !utils.Contains(config.ValueOf.AllowedUsers, chatId) {
		ctx.Reply(u, "You are not allowed to use this bot.", nil)
		return dispatcher.EndGroups
	}

	message := `👋 *Welcome to the Telegram File Stream Bot!*

This bot allows you to generate direct streamable links for media files sent via Telegram.
You can upload videos, audios, or documents and instantly get a streaming link.

✅ *Key features:*
- HTTP streaming for video, audio, and files
- Fast, secure, and easy to use
- Simple interface for sharing content

💡 *How to start?*
Just send a file here or type /help for more information.

🙏 *Support me by clicking here:*
[https://yoelmod.blogspot.com/](https://yoelmod.blogspot.com/)`

	ctx.Reply(u, message, &ext.SendOptions{
		ParseMode: "Markdown",
	})

	return dispatcher.EndGroups
}
