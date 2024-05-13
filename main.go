package main

import (
	"database/sql"
	"github.com/bwmarrin/lit"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kkyr/fig"
	tele "gopkg.in/telebot.v3"
	_ "modernc.org/sqlite"
	"strings"
	"time"
)

var (
	token  string
	db     *sql.DB
	bans   map[int]map[int64]bool
	groups map[int64]int
)

func init() {
	var cfg Config

	err := fig.Load(&cfg, fig.File("config.yml"))
	if err != nil {
		panic(err)
	}

	token = cfg.Token

	// Set lit.LogLevel to the given value
	switch strings.ToLower(cfg.LogLevel) {
	case "logwarning", "warning":
		lit.LogLevel = lit.LogWarning

	case "loginformational", "informational":
		lit.LogLevel = lit.LogInformational

	case "logdebug", "debug":
		lit.LogLevel = lit.LogDebug
	}

	db, err = sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		panic(err)
	}

	execQuery(tblGroup, tblBan, idxGroup, idxBan)

	bans = loadBans()
	groups = loadGroups()
}

func main() {
	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		panic(err)
	}

	b.Handle("/ban", ban)
	b.Handle(tele.OnUserJoined, userJoined)

	b.Start()
}

func userJoined(c tele.Context) error {
	if bans[groups[c.Chat().ID]][c.Sender().ID] {
		chatMember := &tele.ChatMember{
			User: c.Sender(),
			Rights: tele.Rights{
				CanSendMessages: false,
			},
			RestrictedUntil: tele.Forever(),
		}

		err := c.Bot().Ban(c.Chat(), chatMember)
		if err != nil {
			lit.Error("Failed to ban user: %v", err)
		}
	}

	return nil
}

func ban(c tele.Context) error {
	if isAdmin(c) {
		message := c.Message()
		if message.ReplyTo == nil {
			lit.Error("No user to ban")
			return nil
		}

		user := message.ReplyTo.Sender
		chatMember := &tele.ChatMember{
			User: user,
			Rights: tele.Rights{
				CanSendMessages: false,
			},
			RestrictedUntil: tele.Forever(),
		}

		// Add the user to the ban list
		err := c.Bot().Ban(c.Chat(), chatMember)
		if err != nil {
			lit.Error("Failed to ban user: %v", err)
		}

		saveBan(user.ID, groups[c.Chat().ID])
	}

	return nil
}
