#!/usr/bin/env bash
set -euo pipefail

# 環境変数の設定例
# export GO_ENV=dev
# export REDIS_HOST=127.0.0.1
# export REDIS_PORT=6379
# export APP_PORT=8080
# export REDIS_TTL=24h
# export VALKEY="<TODO_VALKEY>"

# TODO: 実環境の値を設定してください
REGION="<TODO_REGION>"
PROJECT_ID="<TODO_PROJECT_ID>"
REPOSITORY="<TODO_REPOSITORY>"
CLOUD_RUN_REGION="<TODO_CLOUD_RUN_REGION>"
IMAGE_TAG="latest"

IMAGE_URL="${REGION}-docker.pkg.dev/${PROJECT_ID}/${REPOSITORY}/spacebase:${IMAGE_TAG}"

echo "Building Docker image..."
docker build -t "${IMAGE_URL}" .

echo "Pushing image to Artifact Registry..."
docker push "${IMAGE_URL}"

echo "Deploying to Cloud Run..."
gcloud run deploy spacebase \
  --image="${IMAGE_URL}" \
  --platform=managed \
  --region="${CLOUD_RUN_REGION}" \
  --allow-unauthenticated \
  --set-env-vars \
GO_ENV="${GO_ENV}",\
REDIS_HOST="${REDIS_HOST}",\
REDIS_PORT="${REDIS_PORT}",\
APP_PORT="${APP_PORT}",\
REDIS_TTL="${REDIS_TTL}",\
VALKEY="${VALKEY}"

echo "Deployment complete."