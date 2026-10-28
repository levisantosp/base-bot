package listeners

import (
	"fmt"
	"log/slog"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
)

func Ready() bot.EventListener {
	return bot.NewListenerFunc(func(e *events.Ready) {
		slog.Info(fmt.Sprintf("%s is ready!", e.User.Tag()))
	})
}
