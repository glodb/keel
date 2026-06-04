# Ops / Docker Reference

## Build

```bash
sudo docker build -t keel -f dockerfiles/Dockerfile .
```

## Run individual services

```bash
sudo docker run --env-file ~/playpalbackend/env/keel.env \
  --log-opt max-size=10m -d --log-opt max-file=3 \
  --network host --restart unless-stopped \
  --name sso keel -env TEST -con SSOSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env \
  --log-opt max-size=10m -d --log-opt max-file=3 \
  --network host --restart unless-stopped \
  --name otp keel -env TEST -con OTPSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env \
  --log-opt max-size=10m -d --log-opt max-file=3 \
  --network host --restart unless-stopped \
  --name admin keel -env TEST -con ADMINSERVICE
```

## Deploying on a new cluster

(Add cluster-specific notes here.)
