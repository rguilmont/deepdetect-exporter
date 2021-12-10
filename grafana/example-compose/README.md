# Stack example with docker-compose

you can simply run :

```shell
chmod -R 777 grafana # Yeah this is nasty :) Not for production  
docker-compose up -d --build
```

grafana is already configured. Password and login is just admin.

## URL

[deepdetect](http://localhost:8080)
[prometheus](http://localhost:9090)
[deepdetect exporter](http://localhost:8181)
[grafana](http://localhost:3000)

you can now load a model in deepdetect :

```shell
curl -X PUT 'http://localhost:8080/services/imageserv' -d '{
      "description": "image classification service",
      "mllib": "caffe",
      "model": {
          "init": "https://deepdetect.com/models/init/desktop/images/classification/ilsvrc_googlenet.tar.gz",
          "repository": "/opt/models/ilsvrc_googlenet",
      "create_repository": true
      },
      "parameters": {
          "input": {
              "connector": "image"
          }
      },
      "type": "supervised"
  }'
```

And finally send a lot of requests, and look at grafana's dashboard :)

```shell
while true; do
            curl -X POST 'http://localhost:8080/predict' -d '{
              "service": "imageserv",
              "parameters": {
              },
              "data": [
                "https://www.deepdetect.com/img/models/ambulance.jpg"
              ]
            }' | jq .
    done
```