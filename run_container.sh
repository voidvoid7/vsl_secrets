#!/bin/bash
set -e

docker build -t voidvoid7/vsl_secrets .
docker run -p 8080:8080 --rm --name vsl_secrets voidvoid7/vsl_secrets