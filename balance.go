/*
	TODO:
	- floating point errors (±1 cent)
	- table overflow: auto, scroll to bottom
*/
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
	"strings"
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

func fileExists(filename string) bool {
	f, err := os.Open(filename)
	f.Close()
	if os.IsNotExist(err) {
		return false
	}
	checkErr(err)
	return true
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/public/") {
			pathOnDisk := "." + r.URL.Path
			if fileExists(pathOnDisk) {
				http.ServeFile(w, r, pathOnDisk)
				return
			}
		}
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
			plusminus := r.PostForm["plusminus"][0]

			amtFloat, _ := strconv.ParseFloat(amt, 64)
			amount := int64(amtFloat * 100)

			if plusminus == "1" {
				amount *= -1
			}

			if amount != 0 {
				register = append(register, &registerEntry{date, desc, amount})
				checkErr(writeRegister(register))
			}
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
		<title>Transaction Register</title>
		<link rel='stylesheet' href='/public/style.css'>
		<form action="/" method="post">
			<fieldset>
				<legend>new entry</legend>
				<div class='row'>
					<div>
						<label>date</label>
						<input type="text" name="date" disabled value="(now)">
					</div>
					<div>
						<label>description</label>
						<input autofocus type="text" name="desc" value="">
					</div>
					<div>
						<label>amount</label>
						<div class='amount'>
							<label><input type="radio" name="plusminus" value="0">+</label>
							<label><input type="radio" name="plusminus" value="1" checked>-</label>
							<input type="text" name="amt" value="">
						</div>
					</div>
				</div>
				<div class='row'>
					<div>
						<input class="primary" type="submit" value="enter">
					</div>
				</div>
			</fieldset>
		</form>
		<table>
		<thead>
			<tr class="primary">
				<th width='33%' colspan='2' class='left mono'>date</th>
				<th width='33%' class='center mono'>transaction</th>
				<th width='33%' class='right mono'>amount</th>
			</tr>
		</thead>
		<tbody>`)
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
				<td class='center'><a href="/?remove=%d">X</a></td>
				<td>%v</td>
				<td>%s</td>
				<td class='right mono'>%s%d.%02d</td>
			</tr>
			`,
				i,
				entry.Date.Format("2006-01-02 15:04 (Monday)"),
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
		</tbody>
		<tfoot>
			<tr>
				<th colspan='3'></th>
				<th class='primary'>
					<div class='mono total'>
						<span>total:</span>
						<span>%s%d.%02d</span>
					</div>
				</th>
			</tr>
		</tfoot>
		</table>`,
			totalSign,
			total/100,
			total%100,
		)
	})
	log.Println("Listening on :8080...")
	log.Fatalln(http.ListenAndServe(":8080", nil))
}
