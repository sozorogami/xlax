package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var mu sync.Mutex
var count int

func echoString(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello, World!")
}

func counter(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	count++
	fmt.Fprintf(w, "Count %d\n", count)
	mu.Unlock()
}

type Status int

const (
	Unknown Status = iota
	Empty
	Occupied
)

type Shitter struct {
	closedThreshold int
	openThreshold   int
	CurrentStatus   Status
	TimeOfLastUse   string
}

func NewShitter(closed int, open int) *Shitter {
	return &Shitter{closedThreshold: closed, openThreshold: open}
}

var shitters = make(map[string]*Shitter)

func (s Status) String() string {
	switch s {
	case 0:
		return "Unknown"
	case 1:
		return "Empty"
	case 2:
		return "Occupied"
	default:
		panic("Wat?")
	}
}

func isAbout(val int, center int) bool {
	const fuzz int = 100
	return (val > (center - fuzz)) && (val < (center + fuzz))
}

func room(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Serve the resource.
		m := make(map[string]string)
		for key, shitter := range shitters {
			m[key] = shitter.CurrentStatus.String()
		}
		rsp, err := json.Marshal(m)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(rsp)
	case "PUT":
		// Update an existing record.
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(body))
		var m = make(map[string]int)
		err = json.Unmarshal(body, &m)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(w, "OK")
		// calculate new value
		v := m["value"]
		s := shitters["9W Mens"].CurrentStatus
		if v > shitters["9W Mens"].closedThreshold {
			s = Occupied
		} else if v < shitters["9W Mens"].openThreshold {
			s = Empty
		}
		mu.Lock()
		shitters["9W Mens"].CurrentStatus = s
		mu.Unlock()
	case "DELETE":
		// Remove the record.
		http.Error(w, "you can't remove %s: we don't have enough already\n", 405)
	default:
		// Give an error message.
		http.Error(w, "What you really want to do???\n", 405)
	}
}

func main() {
	shitters["9W Mens"] = NewShitter(500, 300)

	http.HandleFunc("/room", room)
	http.HandleFunc("/", echoString)
	http.HandleFunc("/count", counter)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "I'm alive")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))

}
