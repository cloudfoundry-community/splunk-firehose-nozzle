import logging
import logging.config
import os
import pytest
import configparser
from os.path import join
from .lib.helper import get_config_folder

_current_dir = os.path.dirname(os.path.realpath(__file__))


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
    parser.addoption("--release", action="store", default="",
                     help="Specify the release to test.")


@pytest.fixture(scope="class")
def test_env():
    """
    provides the config dict based on default and local properties

    local properties over ride default properties. sensitive information should
    be placed in ``config/local.ini`` and this file should not be version
    controlled.
    """
    config_folder = get_config_folder()

    parser = configparser.ConfigParser()
    test_env_var = "PCF_TEST"
    if test_env_var in os.environ:

        environment = os.environ[test_env_var]

        if environment == "integration":

            parser.read([join(config_folder, 'default.ini'),
                         join(config_folder, 'integration.ini')])

        elif environment == "local":

            parser.read([join(config_folder, 'default.ini'),
                         join(config_folder, 'local.ini')])
        else:

            raise ValueError("unknown environment for environment variable "
                             "PCF_TEST: {}".format(environment))
    else:

        parser.read([join(config_folder, 'default.ini'), join(config_folder, 'local.ini')])
    conf = parser._sections
    return conf
