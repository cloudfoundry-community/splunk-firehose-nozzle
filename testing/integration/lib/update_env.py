#! /usr/bin/env python

from .helper import *
_env_var = os.environ
import time
_path = get_project_folder()


def update_environment_variables(input_dict=None):
    default_dict = \
        {
            'ADD_APP_INFO': 'true',
            'EVENTS': 'ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric',
            'SPLUNK_TOKEN': _env_var.get('SPLUNK_TOKEN'),
            'SPLUNK_HOST': _env_var.get('SPLUNK_HOST'),
            'SPLUNK_INDEX': _env_var.get('SPLUNK_INDEX'),
            'FIREHOSE_SUBSCRIPTION_ID': 'splunk-ci',
            'CLIENT_ID': _env_var.get('CLIENT_ID'),
            'CLIENT_SECRET': _env_var.get('CLIENT_SECRET'),
            'ENABLE_EVENT_TRACING': 'true',
            'SKIP_SSL_VALIDATION_CF': 'true',
            'SKIP_SSL_VALIDATION_SPLUNK': 'true',
            'EXTRA_FIELDS': 'name:update-ci-test'
        }

    if input_dict:
        default_dict.update(input_dict)

    path = os.path.join(get_project_folder(), "env.sh")
    with open(path, 'w') as file:
        file.write('''#! /bin/bash''')
        for key, value in default_dict.items():
            file.write("\nexport {0}={1}".format(key, value))
    time.sleep(2)
