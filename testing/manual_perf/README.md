## PCF Nozzle Performance Tests

#### Configuration
Update `env.sh` with desired configuration. Refer to `env_template.sh` file for reference.
Make sure to use long enough `$NOZZLE_DURATION` for datagen to generate all events.
```source testing/manual_perf/env.sh```
#### Environment Set up
- Log in to PCF 

```cf login --skip-ssl-validation -a $API_ENDPOINT -u $API_USER -p $API_PASSWORD```
#### Running Performance Test
```testing/manual_perf/run_perf_test.sh```