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
    "github.com/ttacon/chalk"
    "github.com/spf13/cobra"
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
    var count int
    var from string

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

    var cmdLoad = &cobra.Command{
        Use:   "load",
        Short: "Load verbs from file",
        Run: func(cmd *cobra.Command, args []string) {
            // total:= getTotalVerbs();
            file, err := ioutil.ReadFile(from)
            chk(err)


            var recs []Verb
            err = json.Unmarshal([]byte(file), &recs)
            chk(err)

            _, err = db.Exec("DELETE FROM verbs")
            chk(err)

            for i := 0; i < len(recs); i++ {
                fmt.Printf("Add new verb: %s %v\n", recs[i].Infinitive, recs[i])
                _, err = insert(recs[i].Infinitive, recs[i].Past_simpe, recs[i].Past_participle, recs[i].Translation)
                chk(err)

            }
        },
    }

    cmdLoad.Flags().StringVarP(&from, "from", "f", "./fixtures/irregulars.json", "Load from file")

    var cmdPlay = &cobra.Command{
        Use:   "play",
        Short: "Check verbs",
        Run: func(cmd *cobra.Command, args []string) {
            verbs, err := getVerbs(count)
            chk(err)

            if len(verbs) == 0 {
                fmt.Println("There no verbs to learn")
                os.Exit(1)
            }

            fmt.Printf("Start with %d words...\n", count)

            rl, err := readline.New("> ")
            if err != nil {
                panic(err)
            }

            defer rl.Close()

            var valid string
            correct, incorrect := 0, 0

            lime := chalk.Green.NewStyle().
            WithBackground(chalk.Black).
            WithTextStyle(chalk.Bold).
            Style

            red := chalk.Red.NewStyle().
            WithBackground(chalk.Black).
            WithTextStyle(chalk.Bold).
            Style

            for v := 0; v < len(verbs); v++ {
                fmt.Printf("%d) %s\n", v+1, verbs[v].Translation)
                line, err := rl.Readline()
                if err != nil { // io.EOF
                    break
                }

                valid = fmt.Sprintf("%s %s %s", verbs[v].Infinitive, verbs[v].Past_simpe, verbs[v].Past_participle)
                if line == valid {
                    correct +=1
                    fmt.Println(lime("\u2713"+" valid"))
                } else {
                    incorrect+=1
                    fmt.Println(red("invalid(" + valid + ")"))
                }
            }

            fmt.Println(lime("----------------------------"))
            fmt.Println(lime(fmt.Sprintf("Correct: %d", correct)))
            fmt.Println(red(fmt.Sprintf("In-correct: %d", incorrect)))
        },
    }

    cmdPlay.Flags().IntVarP(&count, "count", "c", 10, "Verbs to check")

    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(cmdPlay, cmdLoad)
    rootCmd.Execute()
}