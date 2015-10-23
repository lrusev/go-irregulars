package main

import (
    "database/sql"
)

func getTotalVerbs() (int) {
    var total int
    err := db.QueryRow("SELECT count(*) FROM verbs").Scan(&total)
    chk(err)

    return total
}

func insert(inf, simple, participle, trans string) (sql.Result, error) {
    return db.Exec("INSERT INTO verbs VALUES (default, $1, $2, $3, $4)", inf, simple, participle, trans)
}


