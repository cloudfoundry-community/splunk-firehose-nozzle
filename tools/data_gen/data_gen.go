package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type config struct {
	EPS         int64
	TotalEvents int64
}

func getIntEnv(key string, defaultValue int64) int64 {
	valStr := os.Getenv(key)
	if len(valStr) == 0 {
		return defaultValue
	}

	if val, err := strconv.ParseInt(valStr, 10, 64); err == nil {
		return val
	}
	return defaultValue
}

func getConfig() *config {
	c := config{}
	c.EPS = getIntEnv("EPS", 10)
	c.TotalEvents = getIntEnv("TOTAL_EVENTS", 0)
	return &c
}

func main() {
	c := getConfig()

	durationPerEvent := time.Duration(int64(time.Second) / c.EPS)
	ticker := time.NewTicker(durationPerEvent)

	uuid := time.Now().UnixNano()
	totalGen := int64(0)
	for {
		select {
		case <-ticker.C:
			totalGen += 1
			fmt.Printf(`{"annotation": "uuid=%d generate data id=%d", "@timestamp":"2017-07-18T22:48:59.763Z","source_host":"1ajkpfgpagq","file":"Dsc2SubsystemAmqpListner.java","method":"spawnNewSubsystemHandler","level":"INFO","line_number":"101","thread_name":"bundle-97-ActorSystem-akka.actor.default-dispatcher-5","@version":1,"logger_name":"com.proximetry.dsc2.listners.Dsc2SubsystemAmqpListner","message":"blahblah-blah|blahblahblah|dsc2| KeyIdRequest :KeyIdRequest(key:xxxxxxxxxxx, id:-xxxxxxxxxxxxxxxxxxx)","class":"com.proximetry.dsc2.listners.Dsc2SubsystemAmqpListner","mdc":{"bundle.version":"0.0.1.SNAPSHOT","bundle.name":"com.proximetry.dsc2","bundle.id":97}}`, uuid, totalGen)
			fmt.Println("")
		}

		if c.TotalEvents > 0 && totalGen >= c.TotalEvents {
			break
		}
	}
	fmt.Printf("end loop\n")

	taken := time.Now().UnixNano() - uuid

	fmt.Printf("uuid=%d data generation done, taken=%f seconds\n", uuid, float64(taken)/float64(time.Second))
	for {
		fmt.Printf("uuid=%d data generation done, taken=%f seconds\n", uuid, float64(taken)/float64(time.Second))
		time.Sleep(30 * time.Second)
	}
}
