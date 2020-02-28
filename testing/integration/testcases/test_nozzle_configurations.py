import pytest
from testing.integration.lib.splunk_api import *
from testing.integration.lib.json_assert import *


class TestSplunkNozzle():
    @pytest.mark.Critical
    def test_enable_event_tracing_is_true(self, request, test_env):
        test_case = request.node.name
        expected_results = ''
        query = get_search_query(input_disc={

        })
        search_results = send_query_to_splunk(url=test_env[''],
                                              core=query,
                                              token=test_env[''])
        assert_json_contains(expected_results, search_results, "Splunk search results mismatch")

