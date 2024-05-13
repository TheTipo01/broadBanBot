package main

import "github.com/bwmarrin/lit"

const (
	tblGroup = "CREATE TABLE `group` ( `groupID` INT(20) NOT NULL, `setID` INT(11) NOT NULL, PRIMARY KEY (`groupID`, `setID`), INDEX `setID` (`setID` );"
	tblBan   = "CREATE TABLE `ban` ( `userID` INT(20) NOT NULL, `setID` INT(11) NOT NULL, PRIMARY KEY (`userID`, `setID`), INDEX `setID` (`setID`), CONSTRAINT `FK_ban_group` FOREIGN KEY (`setID`) REFERENCES `group` (`setID`) ON UPDATE NO ACTION ON DELETE NO ACTION );\n"
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
