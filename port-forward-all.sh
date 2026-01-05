#!/usr/bin/env bash
set -e

NAMESPACE="keda-temporal"

echo "🔌 Port forwarding Temporal UI (http://localhost:30000)..."
kubectl port-forward -n ${NAMESPACE} svc/temporal-ui 30000:8233 > /dev/null 2>&1 &

echo "🔌 Port forwarding Temporal Go API (http://localhost:8080)..."
kubectl port-forward -n ${NAMESPACE} svc/temporal-go-api 18080:8080 > /dev/null 2>&1 &

echo "🔌 Port forwarding RabbitMQ (http://localhost:15672)..."
kubectl port-forward -n ${NAMESPACE} svc/my-rabbitmq 15672:15672 > /dev/null 2>&1 &

echo "🔌 Port forwarding RabbitMQ (http://localhost:5672)..."
kubectl port-forward -n ${NAMESPACE} svc/my-rabbitmq 5672:5672 > /dev/null 2>&1 &

echo "🔌 Port forwarding Metrics (http://localhost:30001/metrics)..."
kubectl port-forward -n ${NAMESPACE} svc/temporal-metrics 30001:9100 > /dev/null 2>&1 &

echo "✅ All port-forwards running in background."
echo "🧭 Visit:"
echo " - Temporal UI → http://localhost:30000"
echo " - API Health  → http://localhost:18080/health"
echo " - RabbitMQ    → http://localhost:15672"
echo " - Metrics     → http://localhost:30001/metrics"
