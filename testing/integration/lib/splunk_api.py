# TODO: Build functions that handle all the activities communicate with Splunk API
import requests
import time
from urllib3.util.retry import Retry
from requests.adapters import HTTPAdapter

TIMEROUT = 500


class SplunkApi():
    def __init__(self, env, logger):
        # Assign configuration object
        self.env = env
        self.logger = logger

    def check_events_from_splunk(self, query,
                                 start_time="-30m@h",
                                 end_time="now",type="events"):
        '''
        send a search request to splunk and return the events from the result
        '''
        query = self.compose_search_query(query,type)
        try:
            events = self.collect_events(query, start_time, end_time,type)
            return events
        except Exception as e:
            #self.logger.error(e)
            raise Exception(e)

    @staticmethod
    def compose_search_query(query,type="events"):
        if type=="events":
            return "search {}".format(query)
        else:
            return query

    def collect_events(self, query, start_time, end_time,type="events"):
        '''
        Collect events by running the given search query
        @param: query (search query)
        @param: start_time (search start time)
        @param: end_time (search end time)
        returns events
        '''

        search_url = '{0}/services/search/jobs?output_mode=json'.format(self.env["splunk_url"])

        # self.logger.info('requesting: %s', search_url)
        data = {
            'search': query,
            'earliest_time': start_time,
            'latest_time': end_time,
        }

        create_job = self.requests_retry_session().post(
            search_url,
            auth=(self.env["splunk_user"], self.env["splunk_password"]),
            verify=False, data=data)
        self.check_request_status(create_job)

        json_res = create_job.json()
        job_id = json_res['sid']
        events = self.wait_for_job_and_get_events(job_id,type)

        return events

    def wait_for_job_and_get_events(self, job_id,type="events"):
        '''
        Wait for the search job to finish and collect the result events
        @param: job_id
        returns events
        '''
        events = []
        job_url = '{0}/services/search/jobs/{1}?output_mode=json'.format(
            self.env["splunk_url"], str(job_id))
        # self.logger.info('requesting: %s', job_url)

        for _ in range(TIMEROUT):
            res = self.requests_retry_session().get(
                job_url,
                auth=(self.env["splunk_user"], self.env["splunk_password"]),
                verify=False)
            self.check_request_status(res)

            job_res = res.json()
            dispatch_state = job_res['entry'][0]['content']['dispatchState']

            if dispatch_state == 'DONE':
                events = self.get_events(job_id,type)
                break
            if dispatch_state == 'FAILED':
                # self.logger.error(f'Search job: {job_url} failed')
                raise Exception(f'Search job: {job_url} failed')

            time.sleep(1)

        return events

    def get_events(self, job_id,type="events"):
        '''
        collect the result events from a search job
        @param: job_id
        returns events
        '''
        event_url = '{0}/services/search/jobs/{1}/{2}?output_mode=json'.format(
            self.env["splunk_url"], str(job_id),type)
        # self.logger.info('requesting: %s', event_url)

        event_job = self.requests_retry_session().get(
            event_url, auth=(self.env["splunk_user"], self.env["splunk_password"]),
            verify=False)
        self.check_request_status(event_job)

        events = event_job.json()['results']

        return events

    def check_request_status(self, req_obj):
        '''
        check if a request is successful
        @param: req_obj
        returns True/False
        '''
        if not req_obj.ok:
            # self.logger.error(f'status code: {str(req_obj.status_code)} \n details: {req_obj.text}')
            raise Exception(f'status code: {str(req_obj.status_code)} \n details: {req_obj.text}')

    def requests_retry_session(self,
                               backoff_factor=0.1,
                               status_forcelist=(500, 502, 504)):
        '''
        create a retry session for HTTP/HTTPS requests
        @param: retries (num of retry time)
        @param: backoff_factor
        @param: status_forcelist (list of error status code to trigger retry)
        @param: session
        returns: session
        '''
        session = requests.Session()
        retry = Retry(
            total=int(self.env['max_retries']),
            backoff_factor=backoff_factor,
            status_forcelist=status_forcelist,
        )
        adapter = HTTPAdapter(max_retries=retry)
        session.mount('http://', adapter)
        session.mount('https://', adapter)

        return session
