# Common functions used in this project

import os


def get_project_folder():
    """
    returns the project root folder
    """
    return os.path.abspath(os.path.join(os.path.dirname(__file__), os.pardir))


def get_config_folder():
    """
    returns the config folder
    """
    return os.path.join(get_project_folder(), "config")



