# kafka-producer-proxy

A flexible and scalable REST API that publishes to configured Kafka brokers.

## Configuration

### `app-config.json`

```json
{
  "debug": false,
  "serverPort": 39000,
  "enableApiAuth": false,
  "enableTLS": false,
  "tlsCert": "",
  "tlsKey": "",
  "useKafkaCertAuth": false,
  "kafkaBrokerGroups": [
    "kafka-cl01"
  ]
}
```

- `debug`             Used for verbose log output, it will be noisy if set to true.
- `serverPort`        The port to expose the API.
- `enableApiAuth`     If true, the header `X-API-TOKEN` must be provided and match the value from the `apiToken` field supplied by the [secrets.json](https://github.com/traviisd/kafka-producer-proxy#secrets-json) file.
- `enableTLS`         If true, serve via https.
- `tlsCert`           Requred if `enableTLS` == true
- `tlsKey`            Required if `enableTLS` == true
- `useKafkaCertAuth`  If true, an [internal-ca.json](https://github.com/traviisd/kafka-producer-proxy#internal-ca-json) must contain the valid certificate details to authenticate to Kafka. [Encryption and Authentication with SSL](https://docs.confluent.io/platform/current/kafka/authentication_ssl.html) 
- `kafkaBrokerGroups` A list of broker mappings. These names must match the keys within `kafkaSecrets` section of the [secrets.json](https://github.com/traviisd/kafka-producer-proxy#secrets-json), e.g. `kafkaSecrets["kafka-cl01"]`.


### `secrets.json`

__NOTE:__ If the cluster certs are internally signed, the environment variable `KAFKA_PRODUCER_PROXY_SSL_CA_LOCATION` must be set to the file or directory path to CA certificate(s) for verifying the broker.

```json
{
  "apiToken": "Generate and place token here",
  "kafkaSecrets": {
    "kafka-cl01": {
      "bootstrap.servers": "broker01:9095,broker02:9095,broker03:9095",
      "enable.idempotence": true,
      "security.protocol": "ssl"
    }
  },
  "oAuthClientSecret": "More to come, unused for now"
}
```

- `apiToken`      The token that allows access to post to the API.
- `kafkaSecrets`  The secrets used to connect to Kafka. It's used a little bit as a key store, but this was the simplest thing that works. 
  - SSL Auth
    ```json
    "KEY": {
      "bootstrap.servers": "broker list",
      "enable.idempotence": true,
      "security.protocol": "ssl"
    }
    ```
  - SASL Auth
    ```json
    "KEY": {
      "bootstrap.servers": "broker list",
      "enable.idempotence": true,
      "security.protocol": "sasl_ssl",
      "sasl.mechanisms": "scram-sha-256",
      "sasl.username":"",
      "sasl.password":""
    }
    ```

### `internal-ca.json`

This only needs to exist if `enableKafkaCertAuth`. This structure is based off of what Hashicorp Vault's [Generate Certificate](https://www.vaultproject.io/api/secret/pki#generate-certificate) 

```json
{
  "ca_chain": [
    "",
  ],
  "certificate": "",
  "expiration": 1636563397,
  "issuing_ca": "",
  "private_key": "",
  "private_key_type": "rsa",
  "serial_number": ""
}
```

## Helm

[Helm Chart](.helm/)

GitHub pages serves as `helm repo add traviisd https://traviisd.github.io/kafka-producer-proxy`
`helm upgrade --install kafka-producer-proxy traviisd/kafka-producer-proxy --version 1.0.0`

### values.yaml

```yaml
# The configuration file app-config.json
appConfig: |-
  {
    "debug": true,
    "serverPort": 39000,
    "enableApiAuth": false,
    "enableTLS": false,
    "tlsCert": "",
    "tlsKey": "",
    "useKafkaCertAuth": false,
    "kafkaBrokerGroups": [
      "kafka-cl01"
    ]
  }

autoscaling:
  minReplicas: 1
  maxReplicas: 10
  targetCPUUtilizationPercentage: 90
  targetMemoryUtilizationPercentage: 90

image:
  repository: traviisd/kafka-producer-proxy
  pullPolicy: IfNotPresent
  tag: 1.0.33

imagePullSecrets: {}
# - name: custom-configuration

labels: {}

service: {}
  # type: ClusterIP
  # port: 80
  # containerPort: 39000
  # annotations: {}

serviceAccount:
  create: true
  # Annotations to add to the service account
  annotations: {}
  # The name of the service account to use.
  # If not set and create is true, a name is generated using the .Release.Name
  name: ""

nodeSelector: {}
# key: value

volumes: []

volumeMounts: []

containers: []

affinity: {}

tolerations: {}

podAnnotations: {}
  # https://www.vaultproject.io/docs/platform/k8s/injector
  # vault.hashicorp.com/agent-limits-cpu: "250m"
  # vault.hashicorp.com/agent-requests-cpu: "10m"
  # vault.hashicorp.com/agent-inject: "true"
  # vault.hashicorp.com/agent-inject-token: "true"
  # vault.hashicorp.com/tls-secret: "tls-crt"
  # vault.hashicorp.com/ca-cert: "/vault/tls/tls.crt"
  # vault.hashicorp.com/secret-volume-path: "/secrets"
  # vault.hashicorp.com/tls-server-name: ""
  # vault.hashicorp.com/service: ""
  # vault.hashicorp.com/role: "kafka-producer-proxy"
  # vault.hashicorp.com/auth-path: "auth/kubernetes"
  # vault.hashicorp.com/agent-inject-secret-secrets.json: "secret/kafka-producer-proxy"
  # vault.hashicorp.com/agent-inject-template-secrets.json: |-
  #   {{- with secret "secret/kafka-producer-proxy" }}{{ .Data.data | toJSONPretty }}{{- end }}
  # vault.hashicorp.com/agent-inject-secret-internal-ca.json: "pki-internal/issue/kafka-producer-proxy"
  # vault.hashicorp.com/agent-inject-template-internal-ca.json: |-
  #   {{ with secret "pki-internal/issue/kafka-producer-proxy" "common_name=kafka-producer-proxy.pki" }}{{ .Data | toJSONPretty }}{{ end }}

ingress: {}
  # annotations: {}
  #   # kubernetes.io/ingress.class: nginx
  # host: kafka-producer-proxy.example.com
  # tlsSecretName: example-name-tls
```
