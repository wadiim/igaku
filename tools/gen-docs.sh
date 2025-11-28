#!/bin/bash

set -xe

common_dirs="commons/dtos,commons/errors,commons/models,commons/utils"
services=(
    "auth-service"
    "visit-service"
    "mail-service"
    "med-service"
    "user-service"
)

for service in "${services[@]}"; do
    swag init -g main.go -o "$service/docs" -d "$service,$common_dirs"
done
