package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/apple/foundationdb/bindings/go/src/fdb"
	"github.com/codahale/hdrhistogram"
)

var (
	frequencySec = flag.Int64("frequency", 1, "reporting frequency")
)

func printLine(f *os.File, args ...interface{}) {
	for i, v := range args {
		if i > 0 {
			f.WriteString("\t")
		}
		fmt.Fprint(f, v)
	}
	f.WriteString("\n")
}

func stats(ms chan metrics, db fdb.Database) {
	freq := time.Duration(*frequencySec) * time.Second
	timer := time.NewTicker(freq).C
	latencyMs := hdrhistogram.New(0, 50000, 3)

	begin := time.Now()

	f := createJournal(db)
	defer f.Close()
	printLine(f, "Seconds", "TxTotal", "TxDelta", "ErrDelta", "Hz", "P50", "P90", "P99", "P999", "P100", "Partitions", "KVTotal", "Disk")

	fmt.Println("     Sec      Hz      Total     Err   P90 ms   P99 ms   MAX ms   Part   KV MiB  Disk MiB")

	var (
		txTotal, txDelta   int64
		errTotal, errDelta int64
	)

	for {
		select {
		case <-timer:

			secTotal := int64(time.Since(begin).Seconds())
			// TODO: account for the delay
			hz := int(txDelta / *frequencySec)

			st, err := getStats(db)

			var kvTotal, partitions, diskTotal int
			if err == nil {
				partitions = st.Cluster.Data.PartitionsCount
				kvTotal = st.Cluster.Data.TotalKvSizeBytes / 1024 / 1024
				diskTotal = st.Cluster.Data.TotalDiskUsedBytes / 1024 / 1024
			} else {
				log.Println(err)
			}

			fmt.Printf("%8d %7d %10d %7d %8d %8d %8d %6d %8d %9d\n",
				secTotal, hz, txTotal, errTotal,
				latencyMs.ValueAtQuantile(90),
				latencyMs.ValueAtQuantile(99),
				latencyMs.ValueAtQuantile(100),
				partitions,
				kvTotal,
				diskTotal,
			)
			printLine(f, secTotal,
				txTotal, txDelta,
				errDelta, hz,
				latencyMs.ValueAtQuantile(50),
				latencyMs.ValueAtQuantile(90),
				latencyMs.ValueAtQuantile(99),
				latencyMs.ValueAtQuantile(99.9),
				latencyMs.ValueAtQuantile(100),
				partitions,
				kvTotal,
				diskTotal,
			)
			// TODO: gather cluster size

			txDelta, errDelta = 0, 0
			latencyMs.Reset()

		case m := <-ms:
			ms := m.nanoseconds / int64(time.Millisecond)
			latencyMs.RecordValue(ms)
			if m.error {
				errDelta++
				errTotal++
			} else {
				txDelta++
				txTotal++
			}

		}
	}

}

func getStats(db fdb.Database) (*AutoGenerated, error) {

	raw, err := db.ReadTransact(func(tr fdb.ReadTransaction) (interface{}, error) {
		return tr.Get(fdb.Key(jsonKey)).Get()
	})

	if err != nil {
		return nil, err
	}

	var gen AutoGenerated
	err = json.Unmarshal(raw.([]byte), &gen)
	if err != nil {
		return nil, err
	}
	return &gen, nil

}

type AutoGenerated struct {
	Client struct {
		Coordinators struct {
			Coordinators []struct {
				Address   string `json:"address"`
				Reachable bool   `json:"reachable"`
			} `json:"coordinators"`
			QuorumReachable bool `json:"quorum_reachable"`
		} `json:"coordinators"`
		DatabaseStatus struct {
			Available bool `json:"available"`
			Healthy   bool `json:"healthy"`
		} `json:"database_status"`
	} `json:"client"`
	Cluster struct {
		Configuration struct {
			CoordinatorsCount int `json:"coordinators_count"`
			Redundancy        struct {
				Factor string `json:"factor"`
			} `json:"redundancy"`
			StorageEngine string `json:"storage_engine"`
			StoragePolicy string `json:"storage_policy"`
			TlogPolicy    string `json:"tlog_policy"`
		} `json:"configuration"`
		Data struct {
			AveragePartitionSizeBytes             int   `json:"average_partition_size_bytes"`
			LeastOperatingSpaceBytesLogServer     int64 `json:"least_operating_space_bytes_log_server"`
			LeastOperatingSpaceBytesStorageServer int64 `json:"least_operating_space_bytes_storage_server"`
			MovingData                            struct {
				InFlightBytes     int `json:"in_flight_bytes"`
				InQueueBytes      int `json:"in_queue_bytes"`
				TotalWrittenBytes int `json:"total_written_bytes"`
			} `json:"moving_data"`
			PartitionsCount int `json:"partitions_count"`
			State           struct {
				Healthy bool   `json:"healthy"`
				Name    string `json:"name"`
			} `json:"state"`
			TotalDiskUsedBytes int `json:"total_disk_used_bytes"`
			TotalKvSizeBytes   int `json:"total_kv_size_bytes"`
		} `json:"data"`
		DatabaseAvailable bool `json:"database_available"`
		DatabaseLocked    bool `json:"database_locked"`
		FaultTolerance    struct {
			MaxMachineFailuresWithoutLosingAvailability int `json:"max_machine_failures_without_losing_availability"`
			MaxMachineFailuresWithoutLosingData         int `json:"max_machine_failures_without_losing_data"`
		} `json:"fault_tolerance"`
		Qos struct {
			LimitingQueueBytesStorageServer int `json:"limiting_queue_bytes_storage_server"`
			LimitingVersionLagStorageServer int `json:"limiting_version_lag_storage_server"`
			PerformanceLimitedBy            struct {
				Description string `json:"description"`
				Name        string `json:"name"`
				ReasonID    int    `json:"reason_id"`
			} `json:"performance_limited_by"`
			ReleasedTransactionsPerSecond float64 `json:"released_transactions_per_second"`
			TransactionsPerSecondLimit    float64 `json:"transactions_per_second_limit"`
			WorstQueueBytesLogServer      int     `json:"worst_queue_bytes_log_server"`
			WorstQueueBytesStorageServer  int     `json:"worst_queue_bytes_storage_server"`
			WorstVersionLagStorageServer  int     `json:"worst_version_lag_storage_server"`
		} `json:"qos"`
		Workload struct {
			Transactions struct {
				Committed struct {
					Counter   int     `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"committed"`
				Conflicted struct {
					Counter   int     `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"conflicted"`
				Started struct {
					Counter   int     `json:"counter"`
					Hz        float64 `json:"hz"`
					Roughness float64 `json:"roughness"`
				} `json:"started"`
			} `json:"transactions"`
		} `json:"workload"`
	} `json:"cluster"`
}
