clean:
	kind delete cluster

kind:
	kind create cluster --config deploy/kind-config.yaml
	kubectl apply -f deploy/device-manager.yaml
	kubectl apply -f deploy/device-plugin.yaml

dep:
	kubectl apply -f deploy/device-manager.yaml
	kubectl apply -f deploy/device-plugin.yaml

bench:
	kubectl apply -f deploy/benchmark.yaml

log:
	kind export logs ./logs

proto:
	protoc \
  --go_out=generated \
  --go_opt=paths=source_relative \
  --go-grpc_out=generated \
  --go-grpc_opt=paths=source_relative \
  device-manager.proto
