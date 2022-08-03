# <img src="img/kubernetes.png" alt="drawing" width="30"/> MonKaos 🐒

> Under development

[![CI](https://github.com/maxgio92/monkaos/actions/workflows/ci.yaml/badge.svg)](https://github.com/maxgio92/monkaos/actions/workflows/ci.yaml)
[![OCI-CI](https://github.com/maxgio92/monkaos/actions/workflows/oci-ci.yaml/badge.svg)](https://github.com/maxgio92/monkaos/actions/workflows/oci-ci.yaml) [![Release](https://github.com/maxgio92/monkaos/actions/workflows/release.yaml/badge.svg)](https://github.com/maxgio92/monkaos/actions/workflows/release.yaml)

Simple [chaos monkey](https://github.com/Netflix/chaosmonkey) on Kubernetes, inspired by [kube-monkey](https://github.com/asobti/kube-monkey).

## Usage

```shell
monkaos [-v <level>] [-c /path/to/config]
```

### Verbosity

One of *debug*, *info*, *warn*, *error*, *fatal*, *panic*.

### Example

```shell
monkaos -c config/samples/config.sample.yaml -v info
```

## Quickstart

### Deploy on Kubernetes

```shell
kubectl apply -k ./deploy
```

#### Try it

```shell
kubectl create deployment --image nginx --replicas=10 nginx
```

## Configuration

The configuration can be specified in yaml or json format.
The expected structure is:

```yaml
scheduler:
  tickPeriodSeconds: 20
  enableRandomLatency: true
  maxLatencySeconds: 1
  deadlineSeconds: 2

chaos:
  terminationGracePeriodSeconds: 10
  victimsPerSchedule: 1
  strategy: RandomPodRandomNamespace
  excludeNamespaces:
  - kube-system
  - kube-public
  - kube-node-lease
  - monkaos-system
```

### Scheduler

The scheduler executes the Schedule that consist of list of Chaos entries.

- `tickPeriodSeconds`: the period every which the scheduler will run the tick, in seconds.
- `enableRandomLatency`: whether to calculate a random latency, with a maximum value of `maxLatencySeconds`.
- `maxLatencySeconds`: the latency that will be added to each scheduled chaos, in seconds. If `enableRandomLatency` is *true*, this is the maximum value allowed for the random latency value.
- `deadlineSeconds`: the deadline period until which a chaos can be completed, in seconds.

### Chaos

Each Chaos entry specifies one or more victims, based on the configured strategy, and the kill time based on the configured scheduler latency.

- `terminationGracePeriodSeconds`: the termination grace period in seconds.
- `victimsPerSchedule`: the number of victims to be selected for chaos, on each schedule execution.
- `strategy`: the strategy to select the victims. Currently the supported strategies are:
  - *RandomPodRandomNamespace*: a random pod is selected from a random namespace, excluding namespaces specified in `excludeNamespaces` array.
- `excludedNamespaces`: the namespaces to exclude from which select the victims.

## Thanks

this project is highly inspired by [kube-chaos](https://github.com/asobti/kube-monkey).
