## Splunk Nozzle

### Setup

The Nozzle requires a uaa user with the scope `doppler.firehose`. One way to create this user
is to add them via the
[uaa.clients](https://github.com/cloudfoundry/uaa-release/blob/master/jobs/uaa/spec)
property in the deployment manifest.

For example:

```
properties:
  uaa:
    clients:
      splunk-firehose-nozzle:
        access-token-validity: 1209600
        authorized-grant-types: authorization_code,client_credentials,refresh_token
        override: true
        secret: <password>
        scope: openid,oauth.approvals,doppler.firehose
        authorities: oauth.login,doppler.firehose
```

### Development

For development against
[bosh-lite](https://github.com/cloudfoundry/bosh-lite),
copy `scripts/dev.sh.template` to `scripts/dev.sh.template` and supply missing values.

Dependency management is done via [Glide](https://glide.sh/)

Reinstall dependencies:
```
glide update
```

Update dependencies (without `strip-vcs` you'll end up with submodules):
```
glide update --strip-vcs
```

Run tests with [Ginkgo](http://onsi.github.io/ginkgo/)
```
ginkgo -r
```

##### CI

https://ci.run-01.haas-26.pez.pivotal.io/pipelines/splunk-firehose-nozzle

### Exploring Events

Here are a few splunk queries to show the events you might want to dashboard

```
index="sandbox" eventType=ValueMetric
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```

```
index="sandbox" eventType=CounterEvent
    | eval job_and_name=source+"-"+name
    | stats values(job_and_name)
```

### Reminder/Todo

- [ ] omitempty on splunk json?
- [ ] Issue w/ Splunk cloud free SSL termination
- [ ] Vendor dependencies
