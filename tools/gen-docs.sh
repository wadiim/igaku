#!/bin/bash

set -xe

common_dirs="commons/dtos,commons/errors,commons/models,commons/utils"
services=(
    "auth-service"
    "geo-service"
    "mail-service"
    "user-service"
    "visit-service"
)

for service in "${services[@]}"; do
    swag init -g main.go -o "$service/docs" -d "$service,$common_dirs"
done
