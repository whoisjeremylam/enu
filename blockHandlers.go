package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/vennd/enu/database"
	"github.com/vennd/enu/enulib"
)

func GetBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	//	 Query DB
	database.Init()
	stmt, err := database.Db.Prepare("select * from blocks order by blockId desc limit 10")
	if err != nil {
		log.Println("Failed to prepare statement. Reason: ")
		panic(err.Error())
	}
	defer stmt.Close()

	//	 Get rows
	var rows *sql.Rows
	rows, err = stmt.Query()
	if err != nil {
		panic(err.Error())
	}

	// Iterate through last 10 blocks and return
	var blocks enulib.Blocks
	i := 1
	for rows.Next() {
		var rowId int64
		var blockId int64
		var status string
		var duration int64

		if err := rows.Scan(&rowId, &blockId, &status, &duration); err != nil {
			log.Fatal(err)
		}

		block := enulib.Block{BlockId: blockId, Status: status, Duration: duration}

		log.Printf("Blockid: %d, Status: %s, Duration: %d\n", block.BlockId, block.Status, block.Duration)

		blocks = append(blocks, block)

		// Maximum of 10 rows
		i++
		if i == 10 {
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(blocks); err != nil {
		panic(err)
	}
}
