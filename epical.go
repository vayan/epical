package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	cookies = new(Jar)
	client  = http.Client{nil, nil, cookies}
	Events  []Event
	port    string
)

type Event struct {
	Scolaryear             string
	Codemodule             string
	Codeinstance           string
	Codeacti               string
	Codeevent              string
	Semester               int
	Instance_location      string
	Titlemodule            string
	Acti_title             string
	Num_event              int
	Start                  string
	End                    string
	Title                  interface{}
	Type_title             string
	Type_code              string
	Nb_hours               string
	Allowed_planning_start string
	Allowed_planning_end   string
	Nb_group               int
	Room                   map[string]interface{}
	Dates                  interface{}
	Module_available       bool
	Module_registered      bool
	Past                   bool
	Allow_register         bool
	Event_registered       interface{}
	Rdv_registered         interface{}
	Allow_token            bool
	Register_student       bool
	Register_prof          bool
	Register_month         bool
	In_more_than_one_month bool
}

type Jar struct {
	cookies []*http.Cookie
}

func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

func auth(login, pass string) {
	ret, err := client.PostForm("https://intra.epitech.eu", url.Values{"login": {login}, "password": {pass}, "remind": {"on"}})
	if ret.StatusCode == 403 || err != nil {
		log.Print(ret.StatusCode, " wrong login or password")
	}
	ret.Body.Close()
}

func json_cal() {
	url := "https://intra.epitech.eu/planning/load?format=json&start=2013-04-08&end=2014-04-14"
	ret, _ := client.Get(url)
	b, _ := ioutil.ReadAll(ret.Body)
	ret.Body.Close()
	re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/") //remove comments in json
	b = re.ReplaceAll(b, nil)
	if err := json.Unmarshal(b, &Events); err != nil {
		log.Print(err)
	}
}

func date_ical(json_date string) string {
	date, _ := time.Parse("2006-01-02 15:04:05", json_date)
	ical_date := date.Format("20060102T150400")
	return ical_date
}

func generate_ical() string {
	json_cal()
	ical := "BEGIN:VCALENDAR\nPRODID:-//Google Inc//Google Calendar 70.9054//EN\nVERSION:2.0\nCALSCALE:GREGORIAN\nMETHOD:PUBLISH\nX-WR-CALNAME:Epitech\nX-WR-CALDESC:"
	for _, val := range Events {
		if val.Event_registered == "registered" {
			startical := date_ical(val.Start)
			endical := date_ical(val.End)
			now := date_ical(time.Now().String())
			ical += "\nBEGIN:VEVENT"
			ical += "\nDTSTART:" + startical
			ical += "\nDTEND:" + endical
			ical += "\nDTSTAMP:" + startical
			ical += "\nUID:" + val.Codeevent
			ical += "\nCREATED:" + startical
			ical += "\nDESCRIPTION:" + val.Type_title
			ical += "\nLAST-MODIFIED:" + now
			if value, ok := val.Room["code"]; ok {
				ical += "\nLOCATION: " + value.(string)
			}
			ical += "\nSEQUENCE:0"
			ical += "\nSTATUS:CONFIRMED"
			ical += "\nSUMMARY:" + val.Acti_title
			ical += "\nTRANSP:OPAQUE"
			ical += "\nEND:VEVENT"
		}
	}
	ical += "\nEND:VCALENDAR"
	return ical
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/calendar")
	fmt.Fprintf(w, generate_ical())
}

func http_server() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":"+port, nil)
}

func main() {
	port = os.Args[3]
	auth(os.Args[1], os.Args[2])
	http_server()
}
