# `otel-wrapper/` — CI-only OTel collector image with env-var config

The upstream `otel/opentelemetry-collector-contrib` image is `FROM scratch` and bakes `--config /etc/otelcol-contrib/config.yaml` into the CMD. GitHub Actions service-containers can't override CMD via `services.<id>.options` (which accepts `docker create` flags only) and can't mount a config file via `services.<id>.volumes` (only host directories, not files). The wrapper exists solely to swap the CMD to `--config=env:OTELCOL_CONFIG` so CI can pass collector config via an env var on the `services:` block.

Dev-host compose YAML uses the upstream image directly with a `command:` field — no wrapper needed there. The wrapper-vs-upstream divergence is one line of CMD; this README exists to make that divergence deliberate rather than incidental.

## Build

```bash
docker build -t galileo-otel-wrapper:ci deploy/compose/otel-wrapper
```

CI builds it as a step before the gateway-integration job runs and references `galileo-otel-wrapper:ci` in `services.otel-collector.image`.
