# Stage 0 OTel collector wrapper.
#
# The upstream otel/opentelemetry-collector-contrib image is FROM scratch,
# so its CMD ["--config", "/etc/otelcol-contrib/config.yaml"] points at a
# baked-in file. GitHub Actions service-containers can't override CMD or
# mount volumes through services.<id>.options. This wrapper exists only
# to redirect CMD to the env-var config provider so CI can pass the
# collector config via OTELCOL_CONFIG.
#
# Compose YAML on the dev host uses the upstream image directly with a
# `command:` field — no wrapper needed there. The wrapper is a CI-only
# artifact; one line of CMD divergence is the acceptable cost.

FROM otel/opentelemetry-collector-contrib:0.152.0
CMD ["--config=env:OTELCOL_CONFIG"]
