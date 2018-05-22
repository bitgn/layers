package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/bitgn/layers/go/benchmark/bench"
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

var (
	journal = flag.String("journal", "", "Journal file name")
)

func createJournal(db fdb.Database, hz int, d *bench.Description, name string) *os.File {

	// experiment journal is a folder that contains following information
	// metadata json
	// tsv rolling log(s)
	// any evidence from the cluster
	// we could probably store that in LMDB, but human-readable is always good

	t := time.Now().UTC()

	folder := *journal

	meta := make(map[string]interface{})

	var (
		info map[string]string
		data []byte
		err  error
	)

	if info, err = loadClusterInfo("/etc/cluster"); info != nil {
		meta["cluster"] = info
	} else {
		fmt.Println("Failed to get cluster info: ", err)
	}

	if len(folder) == 0 {
		folder = t.Format("bench-2006-01-02-15-04-05")

		folder = fmt.Sprintf("%s-%s-%dhz", folder, name, hz)

		if info != nil {
			folder = fmt.Sprintf("%s-%s-%s", folder, info["fdb_type"], info["fdb_count"])
		}

	}

	fmt.Println("Using folder", folder)

	os.MkdirAll(folder, 0755)

	var status []byte

	status, err = getStatus(db)
	if err != nil {
		log.Fatalln("Failed to get FDB status", err)
	}
	statusFile := path.Join(folder, "status.json")
	err = ioutil.WriteFile(statusFile, status, 0644)
	if err != nil {
		log.Fatalln("Failed to dump status.json", err)
	}

	meta["status_file"] = "status.json"
	meta["args"] = os.Args[1:]
	meta["main_tsv"] = "main.tsv"
	meta["time"] = t.Format(time.RFC3339)
	meta["bench-hz"] = hz
	meta["bench-name"] = d.Name
	meta["bench-setup"] = d.Setup
	meta["build"] = buildVersion

	data, err = json.Marshal(meta)
	if err != nil {
		log.Fatalln(err)
	}

	err = ioutil.WriteFile(path.Join(folder, "meta.json"), data, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	var (
		tsvFile *os.File
	)

	tsvFile, err = os.OpenFile(path.Join(folder, "main.tsv"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return tsvFile
}

func getStatus(db fdb.Database) ([]byte, error) {

	raw, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return tr.Get(fdb.Key(jsonKey)).Get()
	})

	if err != nil {
		return nil, err
	}
	return raw.([]byte), nil
}
