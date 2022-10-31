from itertools import count
from attr import fields
import pytest
from lib.json_assert import *
from lib.splunk_api import *
import json


class TestSplunkNozzle():

    @pytest.mark.Critical
    def test_search_event_on_splunk_is_not_empty(self, test_env, splunk_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query="index={}".format(test_env['splunk_index']),
            start_time="-15m@m")

        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={} cf_app_name=data_gen nozzle-event-counter>0",  # nozzle-event-counter should be searchable
        "index={} cf_app_name=data_gen subscription-id::splunk-ci",  # subscription-id should be searchable
        "index={} cf_app_name=data_gen uuid::*"  # uuid should be searchable
    ])
    def test_enable_event_tracing_is_true(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index=*.*",  # wrong splunk index format should return 0 result
        "index=wrong_index"  # wrong splunk index value should return 0 result
    ])
    def test_search_by_incorrect_splunk_index(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input,
            start_time="-15m@m")
        assert len(search_results) == 0, \
            '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={0} cf_space_id=*",  # when cf_space_id is not empty, cf_org_name is searchable
        "index={0} cf_org_id=*",  # cf_org_id add_app_info is not empty, cf_org_name is searchable
        "index={0} cf_org_name=*",  # when add_app_info is not empty, cf_org_name is searchable
        "index={0} cf_space_name=*"  # when add_app_info is not empty, cf_space_name is searchable
    ])
    def test_add_app_info_is_not_empty(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={}| spath event_type | search event_type=LogMessage",
        "index={}| spath event_type | search event_type=ValueMetric",
        "index={}| spath event_type | search event_type=CounterEvent",
        "index={}| spath event_type | search event_type=HttpStartStop",
        "index={}| spath event_type | search event_type=ContainerMetric"
    ])
    def test_search_by_event_type(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={} name::update-ci-test"
    ])
    def test_search_by_extra_fields(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={} arch::old",
        "index={} arch::*"
    ])
    def test_search_by_wrong_extra_fields(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) == 0, \
            '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={} cf_app_name=data_gen subscription-id::* event_type=LogMessage"
    ])
    def test_fields_and_values_in_splunk_event(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        last_event = search_results[0]

        expect_content = {
            '_sourcetype': 'cf:logmessage',
            'cf_app_name': 'data_gen',
            'index': test_env['splunk_index'],
            'source': 'compute',
            'sourcetype': 'cf:logmessage',
            'subscription-id': 'splunk-ci'
        }

        assert_json_contains(expect_content, last_event, "Event raw data results mismatch")
        last_event_raw = json.loads(last_event['_raw'])

        expect_raw_data = {
            "cf_app_name": "data_gen",
            "cf_org_name": "splunk-ci-org",
            "cf_space_name": "splunk-ci-space",
            "event_type": "LogMessage"
        }
        assert_json_contains(expect_raw_data, last_event_raw, "Event raw data results mismatch")


    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "| mstats count where index={} metric_name=*"
    ])
    def test_metric_ingested_data(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_metric_index']),
            start_time="-15m@m",type="results")
        count =   json.loads(json.dumps(search_results[0]))
        assert  int(count['count']) >0
        assert len(search_results) > 0