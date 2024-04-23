# uptime-operator

[![Build](https://github.com/PDOK/uptime-operator/actions/workflows/build-and-publish-image.yml/badge.svg)](https://github.com/PDOK/uptime-operator/actions/workflows/build-and-publish-image.yml)
[![Lint (go)](https://github.com/PDOK/uptime-operator/actions/workflows/lint-go.yml/badge.svg)](https://github.com/PDOK/uptime-operator/actions/workflows/lint-go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/PDOK/uptime-operator)](https://goreportcard.com/report/github.com/PDOK/uptime-operator)
[![Coverage (go)](https://github.com/PDOK/uptime-operator/wiki/coverage.svg)](https://raw.githack.com/wiki/PDOK/uptime-operator/coverage.html)
[![GitHub license](https://img.shields.io/github/license/PDOK/uptime-operator)](https://github.com/PDOK/uptime-operator/blob/master/LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/pdok/uptime-operator.svg)](https://hub.docker.com/r/pdok/uptime-operator)

Kubernetes Operator to watch [Traefik](https://github.com/traefik/traefik) IngressRoute(s) and register these with a (SaaS) uptime monitoring provider.

## Annotations

Traefik `IngressRoute` resources should be annotated in order to successfully register an uptime check. For example:

```yaml
apiVersion: traefik.containo.us/v1alpha1
kind: IngressRoute
metadata:
  name: my-sweet-route
  annotations:
    uptime.pdok.nl/id: "Random string to uniquely identify this check with the provider"
    uptime.pdok.nl/name: "Logical name of the check"
    uptime.pdok.nl/url: "https://site.example/service/wms/v1_0"
    uptime.pdok.nl/tags: "metadata,separated,by,commas"
    uptime.pdok.nl/request-headers: "Accept: application/json, Accept-Language: en"
    uptime.pdok.nl/response-check-for-string-contains: "200 OK"
    uptime.pdok.nl/response-check-for-string-not-contains: "NullPointerException"
```

The `id`, `name` and `url` annotations are mandatory, the rest is optional.

Both `traefik.containo.us/v1alpha1` as well as `traefik.io/v1alpha1` resources are supported.

## Run/usage

```shell
go build github.com/PDOK/uptime-operator/cmd -o manager
```

or

```shell
docker build -t pdok/uptime-operator .
```

```text
USAGE:
   <uptime-controller-manager> [OPTIONS]

OPTIONS:
  -enable-http2
        If set, HTTP/2 will be enabled for the metrics and webhook servers.
  -health-probe-bind-address string
        The address the probe endpoint binds to. (default ":8081")
  -kubeconfig string
        Paths to a kubeconfig. Only required if out-of-cluster.
  -leader-elect
        Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.
  -metrics-bind-address string
        The address the metric endpoint binds to. (default ":8080")
  -metrics-secure
        If set the metrics endpoint is served securely.
  -namespace value
        Namespace(s) to watch for changes. Specify this flag multiple times for each namespace to watch. When not provided all namespaces will be watched.
  -slack-channel string
        The Slack Channel ID for posting updates when uptime checks are mutated.
  -slack-token string
        The token required to access the given Slack channel.
  -uptime-provider string
        Name of the (SaaS) uptime monitoring provider to use. (default "mock")
  -zap-devel
        Development Mode defaults(encoder=consoleEncoder,logLevel=Debug,stackTraceLevel=Warn). Production Mode defaults(encoder=jsonEncoder,logLevel=Info,stackTraceLevel=Error) (default true)
  -zap-encoder value
        Zap log encoding (one of 'json' or 'console')
  -zap-log-level value
        Zap Level to configure the verbosity of logging. Can be one of 'debug', 'info', 'error', or any integer value > 0 which corresponds to custom debug levels of increasing verbosity
  -zap-stacktrace-level value
        Zap Level at and above which stacktraces are captured (one of 'info', 'error', 'panic').
  -zap-time-encoding value
        Zap time encoding (one of 'epoch', 'millis', 'nano', 'iso8601', 'rfc3339' or 'rfc3339nano'). Defaults to 'epoch'.
```

## Develop

The project is written in Go and scaffolded with [kubebuilder](https://kubebuilder.io).

### kubebuilder

Read the manual when you want/need to make changes.
E.g. run `make test` before committing.

### Linting

Install [golangci-lint](https://golangci-lint.run/usage/install/) and run `golangci-lint run`
from the root.
(Don't run `make lint`, it uses an old version of golangci-lint.)

## Misc

### How to Contribute

Make a pull request...

### Contact

Contacting the maintainers can be done through the issue tracker.

## License

```text
MIT License

Copyright (c) 2024 Publieke Dienstverlening op de Kaart

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```


