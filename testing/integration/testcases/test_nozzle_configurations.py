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
        "index={} cf_app_name=data_gen subscription-id::splunk-ci"  # subscription-id should be searchable
    ])
    def test_enable_event_tracing_is_true(self, test_env, splunk_logger, query_input):
        self.splunk_api = SplunkApi(test_env, splunk_logger)

        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index']),
            start_time="-15m@m")
        assert len(search_results) > 0, \
            '\nNumber of events from Splunk should not be {}, however the result is {}'.format(0, len(search_results))


