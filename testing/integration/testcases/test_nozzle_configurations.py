import pytest
from lib.json_assert import *
from lib.splunk_api import *
import json


class TestSplunkNozzle():

    @pytest.mark.Critical
    def test_search_event_on_splunk_is_not_empty(self, test_env, splunk_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query='index=main',
            start_time="-1m@m")

        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index=main cf_app_name=data_gen nozzle-event-counter>100",  # nozzle-event-counter should be searchable
        "index=main cf_app_name=data_gen subscription-id::splunk-ci"  # subscription-id should be searchable
    ])
    def test_enable_event_tracing_is_true(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input,
            start_time="-1m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

        last_event = search_results[0]
        actual_json = json.loads(last_event['_raw'])
        expected_content = {"cf_app_name": "data_gen",
                            "cf_ignored_app": False,
                            "cf_org_name": "pcf-testing",
                            "cf_origin": "firehose",
                            "cf_space_name": "pcf-test-space",
                            "event_type": "LogMessage",
                            "job": "compute",
                            "message_type": "OUT",
                            "msg": {
                                "class": "com.proximetry.dsc2.listners.Dsc2SubsystemAmqpListner",
                                "file": "Dsc2SubsystemAmqpListner.java",
                                "level": "INFO",
                                "line_number": "101",
                                "logger_name": "com.proximetry.dsc2.listners.Dsc2SubsystemAmqpListner",
                                "mdc": {
                                    "bundle.id": 97,
                                    "bundle.name": "com.proximetry.dsc2"
                                },
                                "message": "blahblah-blah|blahblahblah|dsc2| KeyIdRequest :KeyIdRequest(key:xxxxxxxxxxx, id:-xxxxxxxxxxxxxxxxxxx)",
                                "method": "spawnNewSubsystemHandler",
                                "source_host": "1ajkpfgpagq",
                                "thread_name": "bundle-97-ActorSystem-akka.actor.default-dispatcher-5"
                            },
                            "origin": "rep",
                            "source_type": "APP/PROC/WEB"
                            }
        assert_json_contains(expected_content, actual_json, "Splunk search results mismatch")
