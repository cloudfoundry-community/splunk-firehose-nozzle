package utils

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
)

func FormatUUID(uuid *events.UUID) string {
	if uuid == nil {
		return ""
	}
	var uuidBytes [16]byte
	binary.LittleEndian.PutUint64(uuidBytes[:8], uuid.GetLow())
	binary.LittleEndian.PutUint64(uuidBytes[8:], uuid.GetHigh())
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuidBytes[0:4], uuidBytes[4:6], uuidBytes[6:8], uuidBytes[8:10], uuidBytes[10:])
}

func ConcatFormat(stringList []string) string {
	r := strings.NewReplacer(".", "_")
	for i, s := range stringList {
		stringList[i] = strings.TrimSpace(r.Replace(s))
	}

	return strings.Join(stringList, ".")
}

// GetHostIPInfo returns hostname and corresponding IP address
// If empty host is passed in, the current hostname and IP of the host will be
// returned. If the IP of the hostname can't be resolved, an empty IP and an
// error will be returned
func GetHostIPInfo(host string) (string, string, error) {
	var hostname string
	var err error

	hostname = host
	if hostname == "" {
		hostname, err = os.Hostname()
		if err != nil {
			return host, "", err
		}
	}

	ipAddresses, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, "", err
	}

	for _, ia := range ipAddresses {
		return hostname, ia.String(), nil
	}

	return hostname, "", nil
}

func NanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.9f", seconds)
}

// ToJSON tries to detect the JSON pattern for msg first, if msg contains JSON pattern either
// a map or an array (for efficiency), it will try to convert msg to a JSON object. If the convertion
// success, a JSON object will be returned. Otherwise the original msg will be returned
// If the msg param doesn't contain the JSON pattern, the msg will be returned directly
func ToJson(msg string) interface{} {
	trimmed := strings.TrimSpace(msg)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		// Probably the msg can be converted to a map JSON object
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(trimmed), &m); err != nil {
			// Failed to convert to JSON object, just return the original msg
			return msg
		}
		return m
	} else if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
		// Probably the msg can be converted to an array JSON object
		var a []interface{}
		if err := json.Unmarshal([]byte(trimmed), &a); err != nil {
			// Failed to convert to JSON object, just return the original msg
			return msg
		}
		return a
	}
	return msg
}
