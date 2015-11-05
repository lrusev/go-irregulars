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
    "strings"
)

const myconf string = ".myverbs"
const pqconf string = ".verbs"

type Verb struct {
    Id              int    `json:"id"`
    Infinitive      string `json:"infinitive"`
    Past_simpe      string `json:"past_simpe"`
    Past_participle string `json:"past_participle"`
    Translation     string `json:"translation"`
    Active          bool `json:"active"`
    Progress        int `json:"progress"`
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

func rightPad(s string, padStr string, overallLen int) string{
    var padCountInt int
    padCountInt = 1 + ((overallLen-len(padStr))/len(padStr))
    var retStr =  s + strings.Repeat(padStr, padCountInt)
    return retStr[:overallLen]
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
            `translation varchar(100) not null,` +
            `active tinyint default 0,`+
            `progress int(11) default 0` +
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

    lime := chalk.Green.NewStyle().
            WithBackground(chalk.Black).
            WithTextStyle(chalk.Bold).
            Style

    red := chalk.Red.NewStyle().
            WithBackground(chalk.Black).
            WithTextStyle(chalk.Bold).
            Style

    yellow := chalk.Yellow.NewStyle().
            WithBackground(chalk.Black).
            WithTextStyle(chalk.Bold).
            Style

    var cmdLoad = &cobra.Command{
        Use:   "load",
        Short: "Load verbs from file",
        Run: func(cmd *cobra.Command, args []string) {
            file, err := ioutil.ReadFile(from)
            chk(err)


            var recs []Verb
            err = json.Unmarshal([]byte(file), &recs)
            chk(err)

            _, err = db.Exec("DELETE FROM verbs")
            chk(err)

            for i := 0; i < len(recs); i++ {
                fmt.Printf("Add new verb: %s %v\n", recs[i].Infinitive, recs[i])
                _, err = insert(recs[i])
                chk(err)

            }
        },
    }

    cmdLoad.Flags().StringVarP(&from, "from", "f", "./fixtures/irregulars.json", "Load from file")

    var cmdLearn = &cobra.Command{
        Use: "learn",
        Short: "Learn new words",
        Run: func(cmd *cobra.Command, args []string) {
            count, err := cmd.Flags().GetInt("count")
            chk(err)

            all, err := cmd.Flags().GetBool("all")
            chk(err)

            if all == true {
                err := db.QueryRow("SELECT count(*) FROM verbs WHERE active=0").Scan(&count)
                chk(err)
            }

            verbs, err := getVerbs(count, 5, true)
            chk(err)

            if len(verbs) == 0 {
                fmt.Println("There no verbs to learn")
                os.Exit(0)
            }

            var verb Verb

            rl, err := readline.New("Remember(y/n) [y]?")
            chk(err)

            defer rl.Close()

            im, ps, pp := 0, 0, 0

            for m:= 0; m < len(verbs); m++ {
                verb = verbs[m]
                if len(verb.Infinitive) > im {
                    im = len(verb.Infinitive)
                }

                if len(verb.Past_simpe) > ps {
                    ps = len(verb.Past_simpe)
                }

                if len(verb.Past_participle) > pp {
                    pp = len(verb.Past_participle)
                }
            }

            for v:=0; v < len(verbs); v++ {
                verb = verbs[v]
                fmt.Println(yellow(rightPad(verb.Infinitive, " ", im)), lime(" | "), yellow(rightPad(verb.Past_simpe, " ", ps)), lime(" | "), yellow(rightPad(verb.Past_participle, " ", pp)),  lime(" | -> [" + verb.Translation + "]"))
                line, err := rl.Readline()
                if err != nil { // io.EOF
                    break
                }

                if line == "" || strings.ToLower(line) == "y" {
                    verb.Active = true;
                    update(verb)
                }
            }
        },
    }

    cmdLearn.Flags().Int("count", 5, "Verbs to learn")
    cmdLearn.Flags().Bool("all", false, "Specify to check all verbs")

    var cmdCheck = &cobra.Command{
        Use:   "check",
        Short: "Check verbs",
        Run: func(cmd *cobra.Command, args []string) {
            all, err := cmd.Flags().GetBool("all")
            chk(err)

            recall, err := cmd.Flags().GetInt("recall")
            chk(err)

            if all == true {
                count = getTotalVerbs(true);
            }

            verbs, err := getVerbs(count, recall, false)
            chk(err)


            if len(verbs) == 0 {
                fmt.Println("There no verbs to check")
                os.Exit(1)
            }

            fmt.Printf("Start with %d words...\n", len(verbs))

            rl, err := readline.New("> ")
            chk(err)

            defer rl.Close()

            var valid string
            correct, incorrect := 0, 0

            for v := 0; v < len(verbs); v++ {
                fmt.Printf("%d) %s\n", v+1, yellow(verbs[v].Translation))
                line, err := rl.Readline()
                if err != nil { // io.EOF
                    break
                }

                valid = fmt.Sprintf("%s %s %s", verbs[v].Infinitive, verbs[v].Past_simpe, verbs[v].Past_participle)
                if line == valid {
                    correct +=1
                    if verbs[v].Progress < 10 {
                        _, err = progress(verbs[v], 1)
                        chk(err)
                    }

                    fmt.Println(lime("\u2713"+" valid"))
                } else {
                    if verbs[v].Progress > 0 {
                        _, err = progress(verbs[v], -1)
                        chk(err)
                    }

                    incorrect+=1
                    fmt.Println(red("invalid(" + valid + ")"))
                }
            }

            percent := int(float64(correct)/float64(len(verbs))* 100)
            status := "You can do better!"

            switch {
            case percent == 100:
                status = "Excelent!"
            case percent > 85:
                status = "Good!"
            case percent >= 75:
                status = "Not bad!"
            }

            fmt.Println(lime("----------------------------"))
            fmt.Print(lime(fmt.Sprintf("%s | Correct: %d", status, correct)))

            if incorrect > 0 {
                fmt.Print(lime(" | "), red(fmt.Sprintf("Fails: %d", incorrect)))
            }

            fmt.Println()
        },
    }

    cmdCheck.Flags().Int("recall", 5, "Verbs to recall from the end of list")
    cmdCheck.Flags().IntVarP(&count, "count", "c", 10, "Verbs to check")
    cmdCheck.Flags().Bool("all", false, "Specify to check all verbs")
    cmdCheck.Flags().Lookup("all").NoOptDefVal = "true"

    var rootCmd = &cobra.Command{Use: "app"}
    rootCmd.AddCommand(cmdCheck, cmdLoad, cmdLearn)
    rootCmd.Execute()
}