import subprocess
import argparse
import time
import os


nozzle_perf_params = [
    {"hec-workers": "1", "hec-batch-size": "100"},
    {"hec-workers": "2", "hec-batch-size": "100"},
    {"hec-workers": "4", "hec-batch-size": "100"},
    {"hec-workers": "8", "hec-batch-size": "100"},
    {"hec-workers": "16", "hec-batch-size": "100"},
    {"hec-workers": "32", "hec-batch-size": "100"},
    {"hec-workers": "1", "hec-batch-size": "1000"},
    {"hec-workers": "2", "hec-batch-size": "1000"},
    {"hec-workers": "4", "hec-batch-size": "1000"},
    {"hec-workers": "8", "hec-batch-size": "1000"},
    {"hec-workers": "16", "hec-batch-size": "1000"},
    {"hec-workers": "32", "hec-batch-size": "1000"},
]

nozzle_perf_suites = [
    {
        "message-type": "s256byte",
        "extra-fields": "message_type:s256byte",
        "cases": nozzle_perf_params,
    },
    {
        "message-type": "s1kbyte",
        "extra-fields": "message_type:s1kbyte",
        "cases": nozzle_perf_params,
    },
    {
        "message-type": "uns256byte",
        "extra-fields": "message_type:uns256byte",
        "cases": nozzle_perf_params,
    },
    {
        "message-type": "uns1kbyte",
        "extra-fields": "message_type:uns1kbyte",
        "cases": nozzle_perf_params,
    },
]


def run_nozzle_perf(config):
    for suite in nozzle_perf_suites:
        for case in suite["cases"]:
            kvs = ",".join("{}:{}".format(k, v) for k, v in case.iteritems())
            extra_fields = "{},{}".format(suite["extra-fields"], kvs)

            cmd = [
                "./splunk-firehose-nozzle",
                "--api-endpoint", config["api-endpoint"],
                "--user", config["user"],
                "--password", config["password"],
                "--splunk-host", config["splunk-host"],
                "--splunk-token", config["splunk-token"],
                "--splunk-index", config["splunk-index"],
                "--hec-workers", case["hec-workers"],
                "--hec-batch-size", case["hec-batch-size"],
                "--events", "ContainerMetric,CounterEvent,Error,HttpStart,HttpStartStop,HttpStop,LogMessage,ValueMetric",
                "--extra-fields", extra_fields,
                "--add-app-info",
                "--enable-event-tracing",
                "--skip-ssl-validation-cf",
                "--skip-ssl-validation-splunk",
            ]
            print " ".join(cmd)
            execute(cmd, config["duration"])


def execute(cmd, duration):
    has_error = False
    while 1:
        try:
            start = time.time()
            print "start {} {}".format(" ".join(cmd), start)
            out, err = subprocess.Popen(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE).communicate()
        except Exception as e:
            has_error = True
            print e
        else:
            has_error = False
            print "out:", out
            print "err:", err
        finally:
            end = time.time()
            last = end - start
            print "end {} {} duration={}".format(" ".join(cmd), end, last)

            # If we run too short, rerun it
            if last < 0.8 * duration:
                print "run too short, retry..."
                time.sleep(1)
            elif has_error:
                pass
            else:
                break


def run_trafficcontroller(duration):
    for suite in nozzle_perf_suites:
        for case in suite["cases"]:
            ip = "pcf_{}_{}".format(
                suite["message-type"],
                "_".join("{}-{}".format(k, v) for k, v in case.iteritems()))
            cmd = [
                "./trafficcontroller",
                "--config", "loggregator_trafficcontroller.json",
                "--disableAccessControl",
                "--duration", str(duration),
                "--message-type", suite["message-type"],
                "--ip", ip,
            ]

            execute(cmd, duration)
            time.sleep(10)


def main():
    parser = argparse.ArgumentParser(description="Nozzle perf test driver")
    parser.add_argument("--run", dest="run", required=True, help="nozzle or trafficcontroller")
    parser.add_argument("--duration", dest="duration", type=int, required=True, help="how long to run in seconds")
    args = parser.parse_args()

    if args.run == "nozzle":
        config = {
            "api-endpoint": os.environ.get("API_ENDPOINT", "http://trafficcontroller:9911"),
            "user": os.environ.get("API_USER", "admin"),
            "password": os.environ.get("API_PASSWORD", "admin"),
            "splunk-host": os.environ.get("SPLUNK_HOST", "https://heclb1:8088"),
            "splunk-token": os.environ.get("SPLUNK_TOKEN", "00000000-0000-0000-0000-000000000000"),
            "splunk-index": os.environ.get("SPLUNK_INDEX", "main"),
            "duration": args.duration,
        }
        run_nozzle_perf(config)
    else:
        run_trafficcontroller(args.duration)


if __name__ == "__main__":
    main()
