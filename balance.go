package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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

func main() {
	register := []*registerEntry{
		&registerEntry{time.Now(), "initial amount", 69},
		&registerEntry{time.Now(), "deposit", 5000},
		&registerEntry{time.Now(), "melatonin", -1199},
		&registerEntry{time.Now(), "phone", -2694},
	}
	if err := writeRegister(register); err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		io.WriteString(w, `
		<form action="/" method="post">
			<fieldset>
				<legend>new entry</legend>
				<div>
					<label>date</label>
					<input type="text" disabled value="(now)">
				</div>
				<div>
					<label>description</label>
					<input type="text" value="">
				</div>
				<div>
					<label>amount</label>
					<input type="text" value="">
				</div>
				<input type="submit" value="enter">
			</fieldset>
		</form>
		<hr>
		<table border=1>
			<tr>
				<th>date</th>
				<th>transaction</th>
				<th>amount</th>
			</tr>`)

		register, err := getRegister()
		if err != nil {
			fmt.Fprintln(w, "ERROR:", err)
			return
		}
		var total int64 = 0
		for _, entry := range register {
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
				<td>%v</td>
				<td>%s</td>
				<td>%s%d.%02d</td>
			</tr>
			`,
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
				<th colspan=2>total</th>
				<th>%s%d.%02d</th>
			</tr>
		</table>`,
			totalSign,
			total/100,
			total%100,
		)
	})
	http.ListenAndServe(":8080", nil)
}
