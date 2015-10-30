package main

import (
    "database/sql"
)

func getVerbs(count int) ([]Verb, error) {
    var rows *sql.Rows
    var err error
    
    if driver == "mysql" {
        rows, err = db.Query("SELECT * FROM verbs WHERE active=1 ORDER BY RAND() LIMIT ?", count)
    } else {
        rows, err = db.Query("SELECT * FROM verbs WHERE active=1 ORDER BY RANDOM() LIMIT $1", count)
    }
    if err != nil {
        return nil, err
    }

    defer rows.Close()

    var rs = make([]Verb, 0)
    var rec Verb
    for rows.Next() {
        if err = rows.Scan(&rec.Id, &rec.Infinitive, &rec.Past_simpe, &rec.Past_participle, &rec.Translation, &rec.Active); err != nil {
            return nil, err
        }

        rs = append(rs, rec)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return rs, nil
}

func getTotalVerbs() (int) {
    var total int
    err := db.QueryRow("SELECT count(*) FROM verbs").Scan(&total)
    chk(err)

    return total
}

func insert(inf, simple, participle, trans string, active bool) (sql.Result, error) {
    if driver == "mysql" {
        return db.Exec("INSERT INTO verbs VALUES (null, ?, ?, ?, ?, ?)", inf, simple, participle, trans, active)
    } else {
        return db.Exec("INSERT INTO verbs VALUES (default, $1, $2, $3, $4, $5)", inf, simple, participle, trans, active)
    }
}


