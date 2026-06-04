# keel-backend

Backend Microservices<br>

## Running on local machine

Build
sudo docker build -t keel -f dockerfiles/Dockerfile .


sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m  -d --log-opt max-file=3   --network host   --restart unless-stopped   --name sso   keel   -env TEST   -con SSOSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m  -d --log-opt max-file=3   --network host   --restart unless-stopped   --name otp   keel   -env TEST   -con OTPSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m -d  --log-opt max-file=3   --network host   --restart unless-stopped   --name admin   keel   -env TEST   -con ADMINSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m -d  --log-opt max-file=3   --network host   --restart unless-stopped   --name settings   keel   -env TEST   -con EVENTSSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m -d  --log-opt max-file=3   --network host   --restart unless-stopped   --name settings   keel   -env TEST   -con PALSSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m -d  --log-opt max-file=3   --network host   --restart unless-stopped   --name settings   keel   -env TEST   -con PROCESSORSERVICE

sudo docker run --env-file ~/playpalbackend/env/keel.env  --log-opt max-size=10m -d  --log-opt max-file=3   --network host   --restart unless-stopped   --name settings   keel   -env TEST   -con CHATSERVICE

Deploying on new cluster
