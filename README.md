# Elasticsearch-security

## About
This operator provides full lifecycle of Elasticsearch users,roles,rolemapping and alerts.

Mount config file to `/config.yaml` with following parameters:

| Key                  | Value                                                                              |
| -------------------- | ---------------------------------------------------------------------------------- |
| `endpoint`           | Elasticsearch endpoint                                                             |
| `alertAPIPath`       | Path to alerts api endpoint (for example `_opendistro/_alerting/monitors`)         |
| `roleAPIPath`        | Path to roles api endpoint (for example `_opendistro/_security/api/roles`)         |
| `usersAPIPath`       | Path to users api endpoint (for example `_opendistro/_security/api/users`)         |
| `roleMappingAPIPath` | Path to role mappings api endpoint (for example `_opendistro/_security/api/users`) |
| `username`           | User with appropriate permissions                                                  |
| `password`           | User password                                                                      |

## Deploy

Specify configs in `deploy/helm/values.yaml` and deploy with
```bash
helm -n elasticsearch-security-operator upgrade -i elasticsearch-security-operator ./deploy/helm
```
Samples of custom resources can be found in `config/samples`

## TODO:
