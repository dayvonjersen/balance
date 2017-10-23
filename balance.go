package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type registerEntry struct {
	Date        time.Time `json:"date"`
	Description string    `json:"desc"`
	Amount      int64     `json:"amt"`
}

func getRegister() ([]*registerEntry, error) {
	f, err := os.Open("register.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	register := []*registerEntry{}
	err = json.Unmarshal(b, &register)

	return register, err
}

func writeRegister(register []*registerEntry) error {
	f, err := os.Create("register.json")
	if err != nil {
		return err
	}
	b, err := json.Marshal(register)
	if err != nil {
		return err
	}
	io.WriteString(f, string(b))
	return nil
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")

		register, err := getRegister()
		if err != nil {
			fmt.Fprintln(w, "ERROR:", err)
			return
		}

		if r.Method == "POST" {
			checkErr(r.ParseForm())

			date := time.Now()
			desc := r.PostForm["desc"][0]
			amt := r.PostForm["amt"][0]

			amtFloat, _ := strconv.ParseFloat(amt, 64)
			amount := int64(amtFloat * 100)

			register = append(register, &registerEntry{date, desc, amount})
			checkErr(writeRegister(register))
		} else {
			params := r.URL.Query()
			if _, ok := params["remove"]; ok {
				id, _ := strconv.Atoi(params["remove"][0])
				if id < len(register) {
					register = append(register[0:id], register[id+1:]...)
					checkErr(writeRegister(register))
				}
			}
		}

		io.WriteString(w, `
		<form action="/" method="post">
			<fieldset>
				<legend>new entry</legend>
				<div>
					<label>date</label>
					<input type="text" name="date" disabled value="(now)">
				</div>
				<div>
					<label>description</label>
					<input type="text" name="desc" value="">
				</div>
				<div>
					<label>amount</label>
					<input type="text" name="amt" value="">
				</div>
				<input type="submit" value="enter">
			</fieldset>
		</form>
		<hr>
		<table border=1>
			<tr>
				<th>&nbsp;</th>
				<th>date</th>
				<th>transaction</th>
				<th>amount</th>
			</tr>`)

		var total int64 = 0
		for i, entry := range register {
			amt := entry.Amount
			total += amt
			var amtSign string
			switch {
			case amt == 0:
				amtSign = ""
			case amt < 0:
				amtSign = "-"
				amt *= -1
			case entry.Amount > 0:
				amtSign = "+"
			}

			fmt.Fprintf(w, `
			<tr>
				<td><a href="/?remove=%d">X</a></td>
				<td>%v</td>
				<td>%s</td>
				<td>%s%d.%02d</td>
			</tr>
			`,
				i,
				entry.Date,
				entry.Description,
				amtSign,
				amt/100,
				amt%100,
			)

		}
		totalSign := ""
		if total < 0 {
			totalSign = "-"
			total *= -1
		}
		fmt.Fprintf(w, `
			<tr>
				<th colspan=3>total</th>
				<th>%s%d.%02d</th>
			</tr>
		</table>`,
			totalSign,
			total/100,
			total%100,
		)
	})
	log.Println("Listening on :8080...")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
