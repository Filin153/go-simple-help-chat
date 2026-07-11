#!/bin/sh

until mc alias set local \
    "$MINIO_URL" \
    "$MINIO_ROOT_USER" \
    "$MINIO_ROOT_PASSWORD"; do
    sleep 1
done

mc mb --ignore-existing "local/$MINIO_IMAGE_BUCKET"
mc anonymous set download "local/$MINIO_IMAGE_BUCKET"

mc mb --ignore-existing "local/$MINIO_PRIVATE_BUCKET"
