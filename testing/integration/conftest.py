import logging
import logging.config
import os
import sys
import pytest
import configparser
from os.path import join


from lib.helper import get_config_folder, get_project_folder

_current_dir = os.path.dirname(os.path.realpath(__file__))
sys.path.insert(0, get_project_folder())


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
    env_var = os.environ
    parser.addoption("--splunk-url", help="splunk url used to send test data to.",
                     default=env_var.get('SPLUNK_URL'))
    parser.addoption("--splunk-user", help="splunk username",
                     default=env_var.get('SPLUNK_USER'))
    parser.addoption("--splunk-password", help="splunk user password",
                     default=env_var.get('SPLUNK_PASSWORD'))
    parser.addoption("--api-endpoint", help="pleasanton cf api endpoint.",
                     default=env_var.get('API_ENDPOINT'))
    parser.addoption("--splunk-index", help="splunk index on hec setting.",
                     default=env_var.get('SPLUNK_INDEX'))


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

    if os.path.exists(join(config_folder, 'local.ini')):
        parser.read([join(config_folder, 'config.ini'),
                     join(config_folder, 'local.ini')])
    else:
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

    return conf
