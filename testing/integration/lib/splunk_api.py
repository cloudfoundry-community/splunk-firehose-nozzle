# TODO: Build functions that handle all the activities communicate with Splunk API


def get_search_query(input_disc=None):
    # TODO: genterate search query for splunk
    return


def send_query_to_splunk(splunk_logger, url=None, core=None, token=None):
    # TODO: send request to splunk and return the response content
    try:
        response_content = None
        return response_content
    except Exception:
        splunk_logger.error()
        raise

