package main

import (
    "database/sql"
    "strconv"
)

func getVerbs(count, recall int, disabled bool) ([]Verb, error) {
    var rows *sql.Rows
    var err error

    if !disabled {
        union := "(SELECT * FROM verbs WHERE active=1 ORDER BY infinitive desc, progress asc LIMIT " + strconv.Itoa(recall) + ") UNION "

        if driver == "mysql" {
            rows, err = db.Query(union + "(SELECT * FROM verbs WHERE active=1 ORDER BY progress LIMIT ?) ORDER BY RAND()", count - recall)
        } else {
            rows, err = db.Query(union + "(SELECT * FROM verbs WHERE active=1 ORDER BY progress LIMIT $1) ORDER BY RANDOM()", count  - recall)
        }

    } else {
        if driver == "mysql" {
            rows, err = db.Query("SELECT * FROM verbs WHERE active=0 ORDER BY infinitive LIMIT ?", count)
        } else {
            rows, err = db.Query("SELECT * FROM verbs WHERE active=0 ORDER BY infinitive LIMIT $1", count)
        }
    }


    if err != nil {
        return nil, err
    }

    defer rows.Close()

    var rs = make([]Verb, 0)
    var rec Verb
    for rows.Next() {
        if err = rows.Scan(&rec.Id, &rec.Infinitive, &rec.Past_simpe, &rec.Past_participle, &rec.Translation, &rec.Active, &rec.Progress); err != nil {
            return nil, err
        }

        rs = append(rs, rec)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return rs, nil
}

func getTotalVerbs(activeOnly bool) (int) {
    var total int
    if activeOnly {
            err := db.QueryRow("SELECT count(*) FROM verbs WHERE active=1").Scan(&total)
            chk(err)
        } else {
            err := db.QueryRow("SELECT count(*) FROM verbs").Scan(&total)
            chk(err)
        }

    return total
}

func insert(verb Verb) (sql.Result, error) {
    if driver == "mysql" {
        return db.Exec("INSERT INTO verbs VALUES (null, ?, ?, ?, ?, ?, ?)", verb.Infinitive, verb.Past_simpe, verb.Past_participle, verb.Translation, verb.Active, verb.Progress)
    } else {
        return db.Exec("INSERT INTO verbs VALUES (default, $1, $2, $3, $4, $5, $6)", verb.Infinitive, verb.Past_simpe, verb.Past_participle, verb.Translation, verb.Active, verb.Progress)
    }
}

func progress(verb Verb, val int) (sql.Result, error) {
    if driver == "mysql" {
            return db.Exec("UPDATE verbs SET progress = progress + ? WHERE id = ?", val, verb.Id)
        } else {
            return db.Exec("UPDATE verbs SET progress = progress + $1 WHERE id = $2", val, verb.Id)
        }
}

func update(verb Verb) (sql.Result, error) {
     if driver == "mysql" {
        return db.Exec("UPDATE verbs SET infinitive = ?, past_simpe = ?, past_participle = ?, active = ?, progress = ?, translation = ?  WHERE id = ?", verb.Infinitive, verb.Past_simpe, verb.Past_participle, verb.Active, verb.Progress, verb.Translation, verb.Id)
    } else {
        return db.Exec("UPDATE verbs SET infinitive = $1, past_simpe = $2, past_participle = $3, active = $4, progress = $5, translation = $6  WHERE id = $", verb.Infinitive, verb.Past_simpe, verb.Past_participle, verb.Active, verb.Progress, verb.Translation, verb.Id)
    }
}


