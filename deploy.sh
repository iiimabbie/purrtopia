#!/bin/bash

echo "清理 Docker build cache..."
docker builder prune -f

echo "部署 discord-bot..."
docker compose --profile prod up -d --build discord-bot

echo "檢查狀態..."
docker logs purrtopia-discord-bot --tail 10

echo "部署完成！"
