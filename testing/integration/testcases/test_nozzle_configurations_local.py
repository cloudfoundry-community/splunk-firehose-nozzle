import pytest
from lib.update_env import *
from lib.splunk_api import *
import string
import random
import subprocess
_tag = ''.join(random.choices(string.ascii_uppercase + string.digits, k=6))


class TestSplunkNozzleLocal():
    @pytest.fixture(scope='class', autouse=True)
    def setup_class(self, test_env, splunk_logger, nozzle_logger, test_setup):
        update_environment_variables(input_dict={'EVENTS': 'LogMessage',
                                                 'EXTRA_FIELDS': 'test_tag:{},test2.0:nozzle2.0'.format(_tag),
                                                 'ENABLE_EVENT_TRACING': False,
                                                 'ADD_APP_INFO': ''}
                                     )

        path = os.path.join(get_integration_folder(), "scripts")
        env_path = get_project_folder()
        time_interval = 20
        cmd = "cd {0}; ./start_nozzle.sh {1} {2}".format(path, env_path, time_interval)
        proc = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE,
                                stderr=subprocess.STDOUT)
        proc.communicate()

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input, is_result_empty", [
        ("index={0} test_tag::{1} event_type=ValueMetric", True),
        ("index={0} test_tag::{1} event_type=CounterEvent", True),
        ("index={0} test_tag::{1} event_type=LogMessage", False),
        ("index={0} test_tag::{1} event_type=HttpStartStop", True),
        ("index={0} test_tag::{1} event_type=ContainerMetric", True)
    ])
    def test_search_event_by_event_type(self, query_input, is_result_empty, test_env, splunk_logger, nozzle_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)
        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index'], _tag),
            start_time="-10m@m")

        if is_result_empty:
            assert len(search_results) == 0, \
                '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))
        else:
            assert len(search_results) > 0, \
                '\nNumber of events from Splunk should not be {}, however the result is {}'\
                    .format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input, is_result_empty", [
        ("index={0} test_tag::{1} test2.0::nozzle2.0", False),
        ("index={0} test_tag::{1} test::nozzle2.0", True),
        ("index={0} test_tag::{1} test2.0::nozzle", True)
    ])
    def test_search_by_extra_fields(self, query_input, is_result_empty, test_env, splunk_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)
        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index'], _tag),
            start_time="-10m@m")

        if is_result_empty:
            assert len(search_results) == 0, \
                '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))
        else:
            assert len(search_results) > 0, \
                '\nNumber of events from Splunk should not be {}, however the result is {}'\
                    .format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={0} test_tag::{1} nozzle-event-counter>0",  # nozzle-event-counter should not be searchable
        "index={0} test_tag::{1} subscription-id::splunk-ci",  # subscription-id should not be searchable
        "index={0} test_tag::{1} uuid::*"  # uuid should not be searchable
    ])
    def test_enable_event_tracing_is_false(self, test_env, query_input, splunk_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)
        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index'], _tag),
            start_time="-10m@m")

        assert len(search_results) == 0, \
            '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))

    @pytest.mark.Critical
    @pytest.mark.parametrize("query_input", [
        "index={0} test_tag::{1} cf_space_id=*",  # when cf_space_id is false, cf_org_name is not searchable
        "index={0} test_tag::{1} cf_org_id=*",  # cf_org_id add_app_info is false, cf_org_name is not searchable
        "index={0} test_tag::{1} cf_org_name=*",  # when add_app_info is false, cf_org_name is not searchable
        "index={0} test_tag::{1} cf_space_name=*"  # when add_app_info is false, cf_space_name is not searchable
    ])
    def test_add_app_info_is_false(self, test_env, query_input, splunk_logger):
        self.splunk_api = SplunkApi(test_env, splunk_logger)
        search_results = self.splunk_api.check_events_from_splunk(
            query=query_input.format(test_env['splunk_index'], _tag),
            start_time="-10m@m")

        assert len(search_results) == 0, \
            '\nNumber of events from Splunk should be {}, however the result is {}'.format(0, len(search_results))
