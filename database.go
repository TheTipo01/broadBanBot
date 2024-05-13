package main

import "github.com/bwmarrin/lit"

const (
	tblGroup = `CREATE TABLE "groups" ( groupID INTEGER NOT NULL, setID INTEGER NOT NULL, PRIMARY KEY (groupID, setID) );`
	tblBan   = `CREATE TABLE bans ( userID INTEGER NOT NULL, setID INTEGER NOT NULL, PRIMARY KEY (userID, setID), FOREIGN KEY (setID) REFERENCES "group" (setID) ON DELETE NO ACTION );`
	idxGroup = `CREATE INDEX idx_group_setID ON "groups"(setID);`
	idxBan   = `CREATE INDEX idx_ban_setID ON bans(setID);`
)

// Executes a simple query
func execQuery(query ...string) {
	for _, q := range query {
		_, err := db.Exec(q)
		if err != nil {
			lit.Error("Error creating table, %s", err)
		}
	}
}

func loadBans() (result map[int]map[int64]bool) {
	// Load bans from the database
	rows, err := db.Query("SELECT * FROM bans")
	if err != nil {
		panic(err)
	}

	result = make(map[int]map[int64]bool)
	var userID int64
	var setID int

	for rows.Next() {
		err = rows.Scan(&userID, &setID)
		if err != nil {
			lit.Error("Error scanning ban: %v", err)
		}

		if result[setID] == nil {
			result[setID] = make(map[int64]bool)
		}

		result[setID][userID] = true
	}

	return
}

func loadGroups() (result map[int64]int) {
	// Load groups from the database
	rows, err := db.Query("SELECT * FROM groups")
	if err != nil {
		panic(err)
	}

	result = make(map[int64]int)
	var chatID int64
	var setID int

	for rows.Next() {
		err = rows.Scan(&chatID, &setID)
		if err != nil {
			lit.Error("Error scanning group: %v", err)
		}

		result[chatID] = setID
	}

	return
}

func saveBan(userID int64, setID int) {
	_, err := db.Exec("INSERT INTO bans (userID, setID) VALUES (?, ?)", userID, setID)
	if err != nil {
		lit.Error("Error saving ban: %v", err)
	}
}
