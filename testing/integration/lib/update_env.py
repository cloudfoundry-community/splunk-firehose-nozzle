#! /usr/bin/env python

from .helper import *
import time
import yaml
import json
import configparser
from os.path import join

_env_var = os.environ
_path = get_project_folder()


def get_default_env():
    default_dict = \
        {
            'ADD_APP_INFO': "AppName,OrgName,OrgGuid,SpaceName,SpaceGuid",
            'API_ENDPOINT': _env_var.get('API_ENDPOINT'),
            'API_USER': _env_var.get('API_USER'),
            'API_PASSWORD': _env_var.get('API_PASSWORD'),
            'EVENTS': 'ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric',
            'SPLUNK_TOKEN': _env_var.get('SPLUNK_TOKEN'),
            'SPLUNK_HOST': _env_var.get('SPLUNK_HOST'),
            'SPLUNK_INDEX': _env_var.get('SPLUNK_INDEX'),
            'FIREHOSE_SUBSCRIPTION_ID': 'splunk-ci',
            'CLIENT_ID': _env_var.get('CLIENT_ID'),
            'CLIENT_SECRET': _env_var.get('CLIENT_SECRET'),
            'ENABLE_EVENT_TRACING': True,
            'SKIP_SSL_VALIDATION_CF': True,
            'SKIP_SSL_VALIDATION_SPLUNK': True,
            'EXTRA_FIELDS': 'name:update-ci-test'
        }
    config_folder = get_config_folder()
    if os.path.exists(join(config_folder, 'env.json')):
        with open(join(config_folder, 'env.json')) as json_file:
            local_env = json.load(json_file)
        default_dict.update(local_env)

    return default_dict


def update_environment_variables(input_dict=None):
    default_env = get_default_env()
    if input_dict:
        default_env.update(input_dict)

    path = os.path.join(get_project_folder(), "env.sh")
    with open(path, 'w') as file:
        file.write('''#! /bin/bash''')
        for key, value in default_env.items():
            file.write("\nexport {0}={1}".format(key, value))
    time.sleep(2)


def update_nozzle_manifest(nozzle_name=None, instances=None, input_dict=None):
    default_env = get_default_env()
    file_name = os.path.join(_path, ".github/workflows/ci_nozzle_manifest.yml")
    stream = open(file_name, 'r')
    config = yaml.load(stream)
    if instances:
        config['applications'][0].update({'instances': instances})
    if nozzle_name:
        config['applications'][0].update({'name': nozzle_name})

    env_var = config['applications'][0]['env']
    env_var.update(default_env)
    if input_dict:
        env_var.update(input_dict)

    with open(file_name, 'w') as yaml_file:
        yaml_file.write(yaml.dump(config, default_flow_style=False))
    time.sleep(0.5)


def update_data_gen_manifest(input_dict=None):
    file_name = os.path.join(_path, ".github/workflows/data_gen_manifest.yml")
    stream = open(file_name, 'r')
    config = yaml.load(stream)

    env_var = config['applications'][0]
    env_var.update(input_dict)

    with open(file_name, 'w') as yaml_file:
        yaml_file.write(yaml.dump(config, default_flow_style=False))
    time.sleep(0.5)