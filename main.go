package main

import (
    "fmt"
    "log"
    "os/user"
    "database/sql"
    "github.com/FogCreek/mini"
    _ "github.com/lib/pq"
    "encoding/json"
    "io/ioutil"
)

type Verb struct {
    Id              int    `json:"id"`
    Infinitive      string `json:"infinitive"`
    Past_simpe      string `json:"past_simpe"`
    Past_participle string `json:"past_participle"`
    Translation     string `json:"translation"`
}

func fatal(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

func chk(err error) {
    if err != nil {
        fatal(err)
    }
}

func params() string {
    u, err := user.Current()
    fatal(err)
    cfg, err := mini.LoadConfiguration(".verbs")
    fatal(err)

    info := fmt.Sprintf("host=%s port=%s dbname=%s "+
        "sslmode=%s user=%s password=%s ",
        cfg.String("host", "127.0.0.1"),
        cfg.String("port", "5432"),
        cfg.String("dbname", u.Username),
        cfg.String("sslmode", "disable"),
        cfg.String("user", u.Username),
        cfg.String("pass", ""),
    )
    return info
}

var db *sql.DB

func main () {
    var err error
    db, err = sql.Open("postgres", params())
    fatal(err)
    defer db.Close()

    _, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
        `verbs("id" SERIAL PRIMARY KEY,` +
        `"infinitive" varchar(100) not null,` +
        `"past_simpe" varchar(100) not null,` +
        `"past_participle" varchar(100) not null,` +
        `"translation" varchar(100) not null` +
        `)`)
    fatal(err)
    total:= getTotalVerbs();
    file, err := ioutil.ReadFile("./fixtures/irregulars.json")
    chk(err)

    // fmt.Printf("%s\n", string(file))

    var recs []Verb
    err = json.Unmarshal([]byte(file), &recs)
    chk(err)

    if len(recs) > total {
        for i := 0; i < len(recs); i++ {
            fmt.Printf("Add new verb: %s\n", recs[i].Infinitive)
            _, err = insert(recs[i].Infinitive, recs[i].Past_simpe, recs[i].Past_participle, recs[i].Translation)
            chk(err)

        }
    }



    verbs, err := getVerbs(10)
    chk(err)

    for v := 0; v < len(verbs); v++ {
        fmt.Printf("%d) %s\n", v+1, verbs[v].Translation)
    }

}