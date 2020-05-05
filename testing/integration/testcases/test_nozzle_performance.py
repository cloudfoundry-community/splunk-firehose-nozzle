import pytest
from lib.update_env import *
from lib.runner import *
import time


class TestNozzlePerformance:

    @pytest.fixture(scope='class', autouse=True)
    def setup_class(self, test_env, test_setup):
        login_pcf(nozzle_logger)

    def teardown_method(self):
        delete_data_gen(name=self.data_gen_name)

    @classmethod
    def teardown_class(cls):
        delete_pcf_org()

    @pytest.mark.Perf_Binary
    @pytest.mark.parametrize("config_input", [
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:01worker100batch'},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:02worker100batch'},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:04worker100batch'},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:08worker100batch'},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:16worker100batch'},
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:01worker1000batch'},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:02worker1000batch'},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:04worker1000batch'},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:08worker1000batch'},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:16worker1000batch'},
    ])
    def test_nozzle_performance_with_local_binary(self, config_input, test_env):
        self.data_gen_name = config_input['EXTRA_FIELDS'].split(':')[1]
        update_data_gen_manifest(input_dict={'name': self.data_gen_name,
                                             'env': {'GOPACKAGENAME': 'main', 'EPS': 500, 'TOTAL_EVENTS': 0}})
        deploy_date_gen_to_pcf()
        update_environment_variables(input_dict=config_input)
        start_local_nozzle_binary(time_interval=1200)

    @pytest.mark.Perf_Romote
    @pytest.mark.parametrize("config_input", [
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_01worker100batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_02worker100batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_04worker100batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_08worker100batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_16worker100batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_01worker1000batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_02worker1000batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_04worker1000batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_08worker1000batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_16worker1000batch', 'CONSUMER_QUEUE_SIZE': 10000},
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_01worker100batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_02worker100batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_04worker100batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_08worker100batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 100, 'EXTRA_FIELDS': 'test_tag:datagen_16worker100batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 1, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_01worker1000batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 2, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_02worker1000batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 4, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_04worker1000batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 8, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_08worker1000batch', 'CONSUMER_QUEUE_SIZE': 100},
        {'HEC_WORKERS': 16, 'HEC_BATCH_SIZE': 1000, 'EXTRA_FIELDS': 'test_tag:datagen_16worker1000batch', 'CONSUMER_QUEUE_SIZE': 100},
    ])
    def test_nozzle_performance_with_nozzle_deployed_to_pcf(self, config_input, test_env):
        update_nozzle_manifest(nozzle_name='splunk-firehose-nozzle-perf', input_dict=config_input)
        deploy_nozzle_to_pcf()
        self.data_gen_name = config_input['EXTRA_FIELDS'].split(':')[1]
        update_data_gen_manifest(input_dict={'name': self.data_gen_name,
                                             'env': {'GOPACKAGENAME': 'main', 'EPS': 50,
                                                     'SPLUNK_INDEX': 'pcfperf',
                                                     'TOTAL_EVENTS': 1000000}})
        deploy_date_gen_to_pcf()
        time.sleep(300)

    @pytest.mark.Perf_Romote
    @pytest.mark.parametrize("nozzle_instance, data_gen_eps", [
        (1, 1000), (None, 2000), (None, 3000), (None, 4000), (None, 5000),
        (2, 1000), (None, 2000), (None, 3000), (None, 4000), (None, 5000),
        (3, 1000), (None, 2000), (None, 3000), (None, 4000), (None, 5000),
        (4, 1000), (None, 2000), (None, 3000), (None, 4000), (None, 5000),
        (5, 1000), (None, 2000), (None, 3000), (None, 4000), (None, 5000)
    ])
    def test_nozzle_performance_with_multi_instances(self, nozzle_instance, data_gen_eps, test_env, nozzle_logger):
        if nozzle_instance:
            update_nozzle_manifest(nozzle_name='splunk-firehose-nozzle-perf', instances=nozzle_instance)
            deploy_nozzle_to_pcf()
        timestamp = int(time.time())
        self.data_gen_name = 'datagen_{}'.format(timestamp)
        events = data_gen_eps*300
        nozzle_logger.info("Running data-gen: {} to generate {} events".format(self.data_gen_name, events))
        update_data_gen_manifest(input_dict={'name': self.data_gen_name,
                                             'env': {'GOPACKAGENAME': 'main', 'EPS': data_gen_eps,
                                                     'SPLUNK_INDEX': 'pcfperf',
                                                     'TOTAL_EVENTS': events}})
        deploy_date_gen_to_pcf()
        wait_until_date_gen_done(name=self.data_gen_name)
        time.sleep(120)
