## PCF Nozzle Integration Tests Environment Setup

#### Setup Cloud Foundry Nozzle
 - Using remote nozzle environment

    Deploy nozzle to Cloud Foundry env via `cf push`
    ```
    cf login --skip-ssl-validation -a https://api.<your cf sys domain> -u <your cf user name> -p <your cf password>
    
    make deploy-nozzle
    ```
 - Run nozzle binary locally

    Create a `env.sh` file to export all configuration env variables locally.
    ``````
    #!/usr/bin/env bash
     
    export SKIP_SSL_VALIDATION_CF=true
    export SKIP_SSL_VALIDATION_SPLUNK=true
    export ADD_APP_INFO=<String item list of metadata fields: AppName,OrgName,OrgGuid,SpaceName,SpaceGuid>
    export API_ENDPOINT=<CF-ENDPOINT>
    export EVENTS=ValueMetric,CounterEvent,Error,LogMessage,HttpStartStop,ContainerMetric
    export SPLUNK_TOKEN=<HEC-TOKEN>
    export SPLUNK_HOST=<HEC-ENDPOINT>
    export SPLUNK_INDEX=<TARGET-SPLUNK-INDEX>
    export FIREHOSE_SUBSCRIPTION_ID=splunk-nozzle
    export FIREHOSE_KEEP_ALIVE=25s
    export CLIENT_ID=<CF CLIENT ID>
    export CLIENT_SECRET=<CF CLIENT SECRET>
    ``````
    Start nozzle binary
    ```
    cf login --skip-ssl-validation -a https://api.<your cf sys domain> -u <your cf user name> -p <your cf password>
    
    source env.sh
    
    ./splunk-firehose-nozzle
    ```
#### Install python 3

    sudo apt-get install python3.11
    sudo apt-get install python3-pip
    
#### Install python virtualenv

    pip3 install virtualenv
    virtualenv venv
    source venv/bin/activate

#### Install Dependencies

    pip3 install -r requirements.txt

#### Run Automation tests

  - Run all the test cases
      ```
      pytest testing/integration/
      ```

  - Run all the critical test cases tagged with Critical
      ```
      pytest testing/integration -v -m Critical
      ```
    
  - Run specific test case
      ```
      pytest testing/integration/test_method.py::test_func
      ```
      or
      ```
      pytest testing/integration/test_method.py::TestClass::test_func
      ```
      or (if test_func is unique)
      ```
      pytest -k test_func
      ```