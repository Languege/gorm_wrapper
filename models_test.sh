#!/usr/bin/env bash
go run test/main.go -dsn="root:root@tcp(127.0.0.1:3306)/shokudo" \
-debug=true \
-redis_ip=127.0.0.1 \
-redis_port=6379 \
-redis_password=SjhkHD3J5k6H8SjSbK3SC