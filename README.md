
Build gRPC
```
protoc \
  --go_out=generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=generated \
  --go-grpc_opt=paths=source_relative \
  device-manager.proto
```

Create kind cluster
```
kind create cluster --config deploy/kind-config.yaml
```

Build docker from root directory of the project:
```
docker build -t zbsss/device-manager -f deploy/docker/device-manager/Dockerfile .
docker push zbsss/device-manager:latest
kubectl apply -f deploy/device-manager.yaml
```

```
docker build -t zbsss/benchmark -f deploy/docker/benchmark/Dockerfile .
docker push zbsss/benchmark:latest
kubectl apply -f deploy/benchmark.yaml
```

```
docker build -t zbsss/device -f deploy/docker/device/Dockerfile .
docker push zbsss/device:latest
kubectl apply -f deploy/benchmark.yaml
```
kubectl apply -f deploy/device.yaml

```
docker build -t zbsss/device-allocator -f deploy/docker/device-allocator/Dockerfile .
docker push zbsss/device-allocator:latest
```


```
docker build -t zbsss/device-plugin -f deploy/docker/device-plugin/Dockerfile .
docker push zbsss/device-plugin:latest
kubectl apply -f deploy/device-plugin.yaml



kubectl describe daemonset device-plugin -n kube-system

kubectl apply -f deploy/device-plugin-test.yaml

```



```
kubectl port-forward svc/device-manager-service 50051:80
grpcurl -plaintext 127.0.0.1:50051 list
grpcurl -plaintext 127.0.0.1:50051 device_manager.DeviceManager/GetToken
```


```
python3 analysis/time_utilization.py data/data-2023-07-13-09-29-09.json
```
