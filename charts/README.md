# answer

An open-source knowledge-based community software. You can use it quickly to build Q&A community for your products, customers, teams, and more.
## Prerequisites

- Kubernetes 1.20+
## Configuration

The following table lists the configurable parameters of the answer chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `replicaCount`  | Number of answer replicas  | `1` |
| `image.repository` | Image repository | `apache/answer` |
| `image.pullPolicy` | Image pull policy | `Always` |
| `image.tag` | Image tag | `latest` |
| `env` | Optional environment variables for answer | `LOG_LEVEL: INFO` |
| `extraContainers` | Optional sidecar containers to run along side answer | `[]` |
| `persistence.enabled` | Enable or disable persistence for the /data volume | `true` |
| `persistence.accessMode` | Specify the access mode of the persistent volume | `ReadWriteOnce` |
| `persistence.size` | The size of the persistent volume | `5Gi` |
| `persistence.annotations` | Annotations to add to the volume claim | `{}` |
| `imagePullSecrets` | Reference to one or more secrets to be used when pulling images | `[]` |
| `nameOverride` | nameOverride replaces the name of the chart in the Chart.yaml file, when this is used to construct Kubernetes object names. |  |
| `fullnameOverride` | fullnameOverride completely replaces the generated name. |  |
| `serviceAccount.create` | If `true`, create a new service account | `true` |
| `serviceAccount.annotations` | Annotations to add to the service account | `{}` |
| `serviceAccount.name` | Service account to be used. If not set and `serviceAccount.create` is `true`, a name is generated using the fullname template |  |
| `podAnnotations` | Annotations to add to the answer pod | `{}` |
| `podSecurityContext` | Security context for the answer pod | `{}` refer to [Default Security Contexts](#default-security-contexts) |
| `securityContext` | Security context for the answer container | `{}` refer to [Default Security Contexts](#default-security-contexts) |
| `service.type` | The type of service to be used | `ClusterIP` |
| `service.port` | The port that the service should listen on for requests. Also used as the container port. | `80` |
| `ingress.enabled` | Enable or disable ingress. | `false` |
| `resources` | CPU/memory resource requests/limits | `{}` |
| `autoscaling.enabled` | Enable or disable pod autoscaling. If enabled, replicas are disabled. | `false` |
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `tolerations` | Node tolerations for pod assignment | `[]` |
| `affinity` | Node affinity for pod assignment | `{}` |

### Default Security Contexts

The default pod-level and container-level security contexts, below, adhere to the [restricted](https://kubernetes.io/docs/concepts/security/pod-security-standards/#restricted) Pod Security Standards policies.

Default pod-level securityContext:
```yaml
runAsNonRoot: true
seccompProfile:
  type: RuntimeDefault
```

Default containerSecurityContext:
```yaml
allowPrivilegeEscalation: false
capabilities:
  drop:
  - ALL
```
### Installing with a Values file

```console
$ helm install answer -f values.yaml .
```
> **Tip**: You can use the default [values.yaml]

### Installing from the OCI registry

The chart is published to the GitHub Container Registry and can be installed directly with Helm 3.8+.

```console
$ helm install answer oci://ghcr.io/apache/charts/answer --version 0.1.0
```

If you want to inspect or tweak values first, pull the chart and install from the unpacked directory:

```console
$ helm pull oci://ghcr.io/apache/charts/answer --version 0.1.0
$ tar -xzf answer-0.1.0.tgz
$ helm install answer ./answer -f values.yaml
```