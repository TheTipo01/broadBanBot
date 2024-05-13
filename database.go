package main

import "github.com/bwmarrin/lit"

const (
	tblGroup = `CREATE TABLE IF NOT EXISTS "groups" ( groupID INTEGER NOT NULL, setID INTEGER NOT NULL, PRIMARY KEY (groupID, setID) );`
	tblBan   = `CREATE TABLE IF NOT EXISTS bans ( userID INTEGER NOT NULL, setID INTEGER NOT NULL, PRIMARY KEY (userID, setID), FOREIGN KEY (setID) REFERENCES "group" (setID) ON DELETE NO ACTION );`
	idxGroup = `CREATE INDEX IF NOT EXISTS idx_group_setID ON "groups"(setID);`
	idxBan   = `CREATE INDEX IF NOT EXISTS idx_ban_setID ON bans(setID);`
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

func loadBans() (bans map[int64]map[int64]bool) {
	// Load bans from the database
	rows, err := db.Query("SELECT * FROM bans")
	if err != nil {
		panic(err)
	}

	bans = make(map[int64]map[int64]bool)
	var userID, setID int64

	for rows.Next() {
		err = rows.Scan(&userID, &setID)
		if err != nil {
			lit.Error("Error scanning ban: %v", err)
		}

		if bans[setID] == nil {
			bans[setID] = make(map[int64]bool)
		}

		bans[setID][userID] = true
	}

	return
}

func loadGroups() (setToGroups map[int64][]int64, groupToSet map[int64]int64) {
	// Load groups from the database
	rows, err := db.Query("SELECT * FROM groups")
	if err != nil {
		panic(err)
	}

	setToGroups = make(map[int64][]int64)
	groupToSet = make(map[int64]int64)
	var chatID, setID int64

	for rows.Next() {
		err = rows.Scan(&chatID, &setID)
		if err != nil {
			lit.Error("Error scanning group: %v", err)
		}

		setToGroups[setID] = append(setToGroups[setID], chatID)
		groupToSet[chatID] = setID
	}

	return
}

func saveBan(userID, setID int64) {
	_, err := db.Exec("INSERT INTO bans (userID, setID) VALUES (?, ?)", userID, setID)
	if err != nil {
		lit.Error("Error saving ban: %v", err)
	}

	// Update the cache
	if bans[setID] == nil {
		bans[setID] = make(map[int64]bool)
	}
	bans[setID][userID] = true
}

func deleteBan(userID, setID int64) {
	_, err := db.Exec("DELETE FROM bans WHERE userID = ? AND setID = ?", userID, setID)
	if err != nil {
		lit.Error("Error deleting ban: %v", err)
	}

	// Update the cache
	if bans[setID] != nil {
		delete(bans[setID], userID)
	}
}
