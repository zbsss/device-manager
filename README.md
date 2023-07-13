
Build gRPC
```
protoc \
  --go_out=generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=generated \
  --go-grpc_opt=paths=source_relative \
  device-manager.proto
```

Build docker from root directory of the project:
```
docker build -t zbsss/device-manager -f deployment/docker/device-manager/Dockerfile .
docker push zbsss/device-manager:latest
kubectl apply -f deployment/deployment.yaml
```
```
docker build -t zbsss/device -f deployment/docker/device/Dockerfile .
docker push zbsss/device:latest
kubectl apply -f deployment/device.yaml
```



```
kubectl port-forward svc/device-manager-service 50051:80
grpcurl -plaintext 127.0.0.1:50051 list
grpcurl -plaintext 127.0.0.1:50051 device_manager.DeviceManager/GetToken
```


```
python3 analysis/time_utilization.py data/data-2023-07-13-09-29-09.json
```
