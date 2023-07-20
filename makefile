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

logs:
	kind export logs | grep -v "Exporting logs for cluster" | xargs code 
