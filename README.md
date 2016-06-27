## PoC Splunk Nozzle

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

### Reminder/Todo

* Never able to generate `events.Envelope_Error` in real cf deploy
* Switch to batching events

### Notes Todo

* https://github.com/cloudfoundry/dropsonde-protocol/tree/master/events

* https://github.com/cloudfoundry-incubator/cf-lager
* https://github.com/pivotal-golang/lager

* https://github.com/cloudfoundry-incubator/datadog-firehose-nozzle
* https://github.com/cloudfoundry-incubator/datadog-firehose-nozzle-release
* https://github.com/cloudfoundry-community/firehose-to-syslog

For release repo, add errand to setup uaa client, see:
https://github.com/cloudfoundry-community/admin-ui-boshrelease/tree/master/jobs/register_admin_ui
