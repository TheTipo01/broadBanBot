package main

import tele "gopkg.in/telebot.v3"

func isAdmin(c tele.Context) bool {
	switch c.Chat().Type {
	case tele.ChatGroup, tele.ChatSuperGroup:
		members, _ := c.Bot().AdminsOf(c.Chat())
		for _, member := range members {
			if member.User.ID == c.Sender().ID {
				return true
			}
		}
	}

	return false
}
