# -----------------------------
# Config
# -----------------------------
CLUSTER_NAME := keda-demo-cluster
NAMESPACE := keda-temporal

.PHONY: clean keda rabbitmq so build-all build-temporal build-goapi build-worker port check watch-hpa up

# -----------------------------
# Clean + recreate cluster
# -----------------------------
clean:
	@echo "🧹 Deleting existing Kind cluster..."
	-kind delete cluster --name $(CLUSTER_NAME)
	@echo "🚀 Creating new Kind cluster..."
	kind create cluster --name $(CLUSTER_NAME)
	@echo "📦 Creating namespace..."
	kubectl create namespace $(NAMESPACE)
	@echo "📥 Installing KEDA..."
	helm repo add kedacore https://kedacore.github.io/charts
	helm repo update
	helm install keda kedacore/keda --namespace $(NAMESPACE)
	@echo "🐇 Applying RabbitMQ..."
	kubectl apply -f keda/rabbitmq.yaml
	@echo "🧠 Building and loading Temporal image..."
	$(MAKE) build-temporal
	@echo "🔧 Building and loading Go API image..."
	$(MAKE) build-goapi
	@echo "⚙️ Building and loading Worker image..."
	$(MAKE) build-worker
	@echo "📈 Applying KEDA ScaledObject..."
	kubectl apply -f keda/temporal-worker-scaledobject.yaml
	@echo "✅ Setup complete!"
	kubectl get pods -n $(NAMESPACE)

# -----------------------------
# Install components
# -----------------------------
keda:
	helm repo add kedacore https://kedacore.github.io/charts
	helm repo update
	helm install keda kedacore/keda --namespace $(NAMESPACE)

rabbitmq:
	kubectl apply -f keda/rabbitmq.yaml

so:
	kubectl apply -f keda/temporal-worker-scaledobject.yaml

# -----------------------------
# Build targets
# -----------------------------
build-all:
	$(MAKE) build-goapi
	$(MAKE) build-worker

build-temporal:
	docker build -t temporal-dev:latest ./temporal
	kind load docker-image temporal-dev:latest --name $(CLUSTER_NAME)
	kubectl apply -f temporal/temporal-deployment.yaml

build-goapi:
	docker build -t temporal-go-api:latest ./go-api
	kind load docker-image temporal-go-api:latest --name $(CLUSTER_NAME)
	kubectl apply -f go-api/go-api-deployment.yaml

build-worker:
	docker build -t temporal-worker:latest ./worker
	kind load docker-image temporal-worker:latest --name $(CLUSTER_NAME)
	kubectl rollout restart deployment temporal-worker -n $(NAMESPACE)
	kubectl apply -f worker/worker-deployment.yaml

# -----------------------------
# Port forwarding
# -----------------------------
port:
	@echo "🔌 Port forwarding Temporal UI (http://localhost:30000)..."
	kubectl port-forward -n $(NAMESPACE) svc/temporal-ui 30000:8233 > /dev/null 2>&1 &
	@echo "🔌 Port forwarding Go API (http://localhost:18080/health)..."
	kubectl port-forward -n $(NAMESPACE) svc/temporal-go-api 18080:8080 > /dev/null 2>&1 &
	@echo "🔌 Port forwarding RabbitMQ (http://localhost:15672)..."
	kubectl port-forward -n $(NAMESPACE) svc/my-rabbitmq 15672:15672 > /dev/null 2>&1 &
	@echo "🔌 Port forwarding Metrics (http://localhost:30001/metrics)..."
	kubectl port-forward -n $(NAMESPACE) svc/temporal-metrics 30001:9100 > /dev/null 2>&1 &
	@echo "✅ All port-forwards running in background."

# -----------------------------
# Verification
# -----------------------------
check:
	kubectl get pods -n $(NAMESPACE)
	@echo ""
	@echo "🌐 Test API health endpoint → curl http://localhost:18080/health"

watch-hpa:
	kubectl get hpa -n $(NAMESPACE) -w

# -----------------------------
# Full bring-up
# -----------------------------
up:
	$(MAKE) clean
	$(MAKE) port
