
# Elasticsearch-security-operator
[![Go Report Card](https://goreportcard.com/badge/github.com/aberestyak/elasticsearch-security-operator)](https://goreportcard.com/report/github.com/aberestyak/elasticsearch-security-operator) [![Go Report Card](https://img.shields.io/docker/image-size/berestyak/elasticsearch-security-operator/0.1.5)

This operator provides full lifecycle of Elasticsearch users,roles,rolemapping and alerts.

## Configuration

You can pass configuration with environment variables or file with following parameters:

| Key                  | Environment variable                 | Value                                                                                     |
| -------------------- | ------------------------------------ | ----------------------------------------------------------------------------------------- |
| `endpoint`           | `ELASTICSEARCH_ENDPOINT`             | Elasticsearch endpoint                                                                    |
| `alertAPIPath`       | `ELASTICSEARCH_ALERT_API_PATH`       | Path to alerts api endpoint (for example `_opendistro/_alerting/monitors`)                |
| `roleAPIPath`        | `ELASTICSEARCH_ROLE_API_PATH`        | Path to roles api endpoint (for example `_opendistro/_security/api/roles`)                |
| `userAPIPath`       | `ELASTICSEARCH_USER_API_PATH`        | Path to users api endpoint (for example `_opendistro/_security/api/internalusers`)        |
| `tenantAPIPath`     | `ELASTICSEARCH_TENANT_API_PATH`      | Path to tenants api endpoint (for example `_opendistro/_security/api/tenants`)            |
| `roleMappingAPIPath` | `ELASTICSEARCH_ROLEMAPPING_API_PATH` | Path to role mappings api endpoint (for example `_opendistro/_security/api/rolesmapping`) |
| `extraCACertFile`    | `EXTRA_CA_CERT_FILE`                 | Path to file with custom CA certificate(s)                                                |
| `username`           | `ELASTICSEARCH_USERNAME`             | User with appropriate permissions                                                         |
| `password`           | `ELASTICSEARCH_PASSWORD`             | User password                                                                             |



## Build

### Requirements

* [gosec](https://github.com/securego/gosec)
* golint
* [golangci-lint](https://github.com/golangci/golangci-lint)
* Installed [envtest](https://book.kubebuilder.io/reference/envtest.html) binaries

Export `VERSION` variable and execute

```bash
make docker-build
```

## Deploy

Specify configs in `deploy/helm/values.yaml` and deploy with
```bash
helm -n elasticsearch-security-operator upgrade -i elasticsearch-security-operator ./deploy/helm
```
Samples of custom resources can be found in `config/samples`

## TODO:

- [ ] Refactor alert controller
- [ ] Refactor controller's methods
