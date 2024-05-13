package main

import (
	"database/sql"
	"github.com/bwmarrin/lit"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kkyr/fig"
	tele "gopkg.in/telebot.v3"
	_ "modernc.org/sqlite"
	"strconv"
	"strings"
	"time"
)

var (
	token       string
	db          *sql.DB
	bans        map[int64]map[int64]bool
	setToGroups map[int64][]int64
	groupToSet  map[int64]int64
	groupsCache map[int64]*tele.Chat
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
	setToGroups, groupToSet = loadGroups()
	groupsCache = make(map[int64]*tele.Chat)
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
	b.Handle(&tele.InlineButton{Unique: "unbanBtn"}, unban)

	lit.Info("broadBanBot ready to exterminate")
	b.Start()
}

func userJoined(c tele.Context) error {
	for _, sender := range c.Message().UsersJoined {
		if bans[groupToSet[c.Chat().ID]][sender.ID] {
			chatMember := &tele.ChatMember{
				User: &sender,
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
	}

	return nil
}

func ban(c tele.Context) error {
	var err error

	if isAdmin(c) {
		message := c.Message()
		if message.ReplyTo == nil {
			lit.Error("No user to ban")
			return nil
		}

		user := message.ReplyTo.Sender
		chatMember := &tele.ChatMember{
			User:            user,
			RestrictedUntil: tele.Forever(),
		}

		// Ban the user from every group within the set
		for _, groupID := range setToGroups[groupToSet[c.Chat().ID]] {
			if _, ok := groupsCache[groupID]; !ok {
				groupsCache[groupID], err = c.Bot().ChatByID(groupID)
				if err != nil {
					lit.Error("Error getting group %d: %v", groupID, err)
				}
			}

			err = c.Bot().Ban(groupsCache[groupID], chatMember)
			if err != nil {
				lit.Error("Error banning user from group %d: %v", groupID, err)
			}
		}

		// Add the user to the ban list
		saveBan(user.ID, groupToSet[c.Chat().ID])

		button := tele.InlineButton{
			Unique: "unbanBtn",
			Text:   "Sbanna",
			Data:   strconv.FormatInt(user.ID, 10) + "|" + strconv.FormatInt(groupToSet[c.Chat().ID], 10),
		}

		markup := &tele.ReplyMarkup{InlineKeyboard: [][]tele.InlineButton{{button}}}

		err = c.Reply("L'utente è stato bannato dal Cosmo di UniTo.", markup)
	}

	return err
}

func unban(c tele.Context) error {
	var err error

	if isAdmin(c) {
		data := strings.Split(c.Callback().Data, "|")
		userID, _ := strconv.ParseInt(data[0], 10, 64)
		setID, _ := strconv.ParseInt(data[1], 10, 64)

		user := &tele.User{ID: userID}

		// Unban the user from every group within the set
		for _, groupID := range setToGroups[setID] {
			if _, ok := groupsCache[groupID]; !ok {
				groupsCache[groupID], err = c.Bot().ChatByID(groupID)
				if err != nil {
					lit.Error("Error getting group %d: %v", groupID, err)
				}
			}

			err = c.Bot().Unban(groupsCache[groupID], user)
			if err != nil {
				lit.Error("Error unbanning user from group %d: %v", groupID, err)
			}
		}

		// Remove the user from the ban list
		deleteBan(userID, setID)

		err = c.Edit("L'utente è stato sbannato dal Cosmo di UniTo.")
	}

	return err
}
