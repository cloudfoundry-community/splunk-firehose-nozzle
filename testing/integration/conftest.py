import logging
import logging.config
import os
import sys
import pytest
import configparser
from os.path import join
import subprocess
import time


from lib.helper import get_config_folder, get_project_folder

_current_dir = os.path.dirname(os.path.realpath(__file__))
sys.path.insert(0, get_project_folder())
_env_var = os.environ


@pytest.fixture(scope="session", autouse=True)
def setup_logging():
    logging.config.fileConfig(os.path.join(_current_dir, "logging.conf"))


@pytest.fixture(scope="session", autouse=True)
def nozzle_logger(setup_logging):
    return logging.getLogger("nozzle")


@pytest.fixture(scope="session", autouse=True)
def splunk_logger(setup_logging):
    return logging.getLogger("splunk")


def pytest_addoption(parser):
    """
    This function is sued to add command line parameters to test suite
    """
    parser.addoption("--splunk-url", help="splunk url used to send test data to.",
                     default=_env_var.get('SPLUNK_URL'))
    parser.addoption("--splunk-user", help="splunk username",
                     default=_env_var.get('SPLUNK_USER'))
    parser.addoption("--splunk-password", help="splunk user password",
                     default=_env_var.get('SPLUNK_PASSWORD'))
    parser.addoption("--api-endpoint", help="pleasanton cf api endpoint.",
                     default=_env_var.get('API_ENDPOINT'))
    parser.addoption("--splunk-index", help="splunk index on hec setting.",
                     default=_env_var.get('SPLUNK_INDEX'))
    parser.addoption("--splunk-metric-index", help="splunk index on hec setting.",
                     default=_env_var.get('SPLUNK_METRIC_INDEX'))


@pytest.fixture(scope="class")
def test_setup(request, nozzle_logger):
    nozzle_logger.info("Stopping nozzle...")
    cmd = "cf login --skip-ssl-validation -a {0} -u {1} -p {2} -o system -s system; " \
          "cf stop splunk-firehose-nozzle".format(_env_var['API_ENDPOINT'],
                                                  _env_var['API_USER'],
                                                  _env_var['API_PASSWORD'])

    proc = subprocess.Popen(cmd, shell=True, stdout=subprocess.PIPE,
                            stderr=subprocess.STDOUT)
    output, error = proc.communicate()
    time.sleep(5)


@pytest.fixture(scope="class")
def test_env(request):
    """
    provides the config dict based on default and local properties

    local properties over ride default properties. sensitive information should
    be placed in ``config/local.ini`` and this file should not be version
    controlled.
    """
    config_folder = get_config_folder()
    parser = configparser.ConfigParser()
    parser.read(join(config_folder, 'config.ini'))

    cfg = parser.items('DEFAULT')
    conf = dict(cfg)

    if request.config.getoption("--splunk-url"):
        conf["splunk_url"] = request.config.getoption("--splunk-url")
    if request.config.getoption("--splunk-user"):
        conf["splunk_user"] = request.config.getoption("--splunk-user")
    if request.config.getoption("--splunk-password"):
        conf["splunk_password"] = request.config.getoption("--splunk-password")
    if request.config.getoption("--api-endpoint"):
        conf["api_endpoint"] = request.config.getoption("--api-endpoint")
    if request.config.getoption("--splunk-index"):
        conf["splunk_index"] = request.config.getoption("--splunk-index")
    if request.config.getoption("--splunk-metric-index"):
        conf["splunk_metric_index"] = request.config.getoption("--splunk-metric-index")
    
    if os.path.exists(join(config_folder, 'local.ini')):
        parser.read(join(config_folder, 'local.ini'))
        conf.update(parser.items('LOCAL'))

    return conf
