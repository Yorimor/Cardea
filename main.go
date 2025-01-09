package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-systemd/v22/dbus"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	targetUnits = make(map[string]string)
	dbusUnits   = []dbus.UnitStatus{}
)

type unitStatus struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	SubStatus string `json:"subStatus"`
}

func main() {
	loadUnits()
	initUnits()
	log.Println("Units initialized, starting HTTP server")

	http.HandleFunc("/", homePage)
	http.HandleFunc("/status/", statusPage)

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func homePage(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s\n", r.Method, r.URL.Path)

	http.ServeFile(w, r, "index.html")
}

func statusPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s - %s\n", r.Method, r.URL.Path)

	j, _ := json.Marshal(getStatus())
	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func loadUnits() {
	jsonFile, err := os.Open("units.json")
	if err != nil {
		panic(err)
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	json.Unmarshal([]byte(byteValue), &targetUnits)
}

func initUnits() {
	ctx := context.Background()

	systemdConnection, err := dbus.NewSystemConnectionContext(ctx)
	if err != nil {
		fmt.Printf("Failed to connect to systemd: %v\n", err)
		panic(err)
	}
	defer systemdConnection.Close()

	listOfUnits, err := systemdConnection.ListUnitsContext(ctx)
	if err != nil {
		fmt.Printf("Failed to list units: %v\n", err)
		panic(err)
	}

	for targetSystemdUnit, _ := range targetUnits {
		for _, unit := range listOfUnits {
			if unit.Name == targetSystemdUnit {
				dbusUnits = append(dbusUnits, unit)
				break
			}
		}
	}
}

func getStatus() map[string]unitStatus {
	initUnits()
	var status = make(map[string]unitStatus)

	for _, unit := range dbusUnits {
		//fmt.Printf("Unit %s is in state %s and substate %s\n", unit.Name, unit.ActiveState, unit.SubState)
		status[unit.Name] = unitStatus{
			Name:      targetUnits[unit.Name],
			Status:    unit.ActiveState,
			SubStatus: unit.SubState,
		}
	}

	return status
}
