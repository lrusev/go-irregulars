package main

import (
    "fmt"
    "log"
    "os"
    "os/user"
    "database/sql"
    "github.com/FogCreek/mini"
    _ "github.com/lib/pq"
    _ "github.com/go-sql-driver/mysql"
    "encoding/json"
    "io/ioutil"
    "gopkg.in/readline.v1"
)

const myconf string = ".myverbs"
const pqconf string = ".verbs"

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

func connect() (*sql.DB, error) {
    var info string

    u, err := user.Current()
    fatal(err)

    if _, err = os.Stat(myconf); os.IsNotExist(err) {
        cfg, err := mini.LoadConfiguration(pqconf)
        fatal(err)

        info = fmt.Sprintf("host=%s port=%s dbname=%s "+
            "sslmode=%s user=%s password=%s ",
            cfg.String("host", "127.0.0.1"),
            cfg.String("port", "5432"),
            cfg.String("dbname", u.Username),
            cfg.String("sslmode", "disable"),
            cfg.String("user", u.Username),
            cfg.String("pass", ""),
        )

        driver = "postgres"
    } else {
        cfg, err := mini.LoadConfiguration(myconf)
        fatal(err)

        driver = "mysql"
        info = fmt.Sprintf("%s:%s@/%s?charset=utf8mb4,utf8&collation=utf8_general_ci",
            cfg.String("user", u.Username),
            cfg.String("pass", ""),
            cfg.String("dbname", u.Username),
        )
    }

    return sql.Open(driver, info)
}

var db *sql.DB
var driver string

func main () {
    var err error

    db, err = connect()
    fatal(err)
    defer db.Close()

    if driver == "mysql" {
        _, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
            `verbs(id int(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,` +
            `infinitive varchar(100) not null,` +
            `past_simpe varchar(100) not null,` +
            `past_participle varchar(100) not null,` +
            `translation varchar(100) not null` +
            `) engine=InnoDb Default charset=utf8 collate=utf8_general_ci`)
    } else {
        _, err = db.Exec("CREATE TABLE IF NOT EXISTS " +
            `verbs("id" SERIAL PRIMARY KEY,` +
            `"infinitive" varchar(100) not null,` +
            `"past_simpe" varchar(100) not null,` +
            `"past_participle" varchar(100) not null,` +
            `"translation" varchar(100) not null` +
            `)`)
    }

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
            fmt.Printf("Add new verb: %s %v\n", recs[i].Infinitive, recs[i])
            _, err = insert(recs[i].Infinitive, recs[i].Past_simpe, recs[i].Past_participle, recs[i].Translation)
            chk(err)

        }
    }

    
    count := 10


    verbs, err := getVerbs(count)
    chk(err)

    fmt.Printf("Start with %d words...\n", count)

    /*for v := 0; v < len(verbs); v++ {
        fmt.Printf("%d) %s\n", v+1, verbs[v].Translation)
    }*/

    rl, err := readline.New("> ")
    if err != nil {
        panic(err)
    }

    defer rl.Close()

    var valid string
    correct, incorrect := 0, 0

    for v := 0; v < len(verbs); v++ {
        fmt.Printf("%d) %s\n", v+1, verbs[v].Translation)
        line, err := rl.Readline()
        if err != nil { // io.EOF
            break
        }

        valid = fmt.Sprintf("%s %s %s", verbs[v].Infinitive, verbs[v].Past_simpe, verbs[v].Past_participle)
        if line == valid {
            correct +=1
            println("valid")
        } else {
            incorrect+=1
            println("invalid(", valid, ")")
        }
    }

    println(fmt.Sprintf("Correct: %d", correct))
    println(fmt.Sprintf("In-correct: %d", incorrect))
}