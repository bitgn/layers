package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
)

var (
	jsonKey = append([]byte{255, 255}, []byte("/status/json")...)
)

func loadClusterInfo(name string) (map[string]string, error) {

	if _, err := os.Stat(name); err != nil {
		return nil, nil
	}

	f, _ := os.Open(name)
	defer f.Close()

	m := make(map[string]string)
	// Create a new Scanner for the file.
	scanner := bufio.NewScanner(f)
	// Loop over all lines in the file and print them.
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])
		m[key] = value
	}
	return m, nil
}

func mustOpenJournal(db fdb.Database) *os.File {

	name := time.Now().Format("2006-01-02-15-04-05")

	status, err := getStatus(db)
	if err != nil {
		log.Fatalln("Failed to get FDB status", err)
	}

	var (
		info map[string]string
		data []byte
	)

	if info, err = loadClusterInfo("/etc/cluster"); info != nil {
		status["cluster"] = info
	}
	status["args"] = os.Args[1:]

	data, err = json.Marshal(status)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(name+".json", data, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	var (
		tsvFile *os.File
	)

	tsvFile, err = os.OpenFile(name+".tsv", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return tsvFile
}

func getStatus(db fdb.Database) (map[string]interface{}, error) {

	raw, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return tr.Get(fdb.Key(jsonKey)).Get()
	})

	if err != nil {
		return nil, err
	}

	m := make(map[string]interface{})
	err = json.Unmarshal(raw.([]byte), &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
