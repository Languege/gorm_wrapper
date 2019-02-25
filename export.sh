#!/usr/bin/env bash

go run gen/main.go  -dsn="root:root@tcp(127.0.0.1:3306)/shokudo" \
-db_name="shokudo" \
-out_dir="models"  \
-package="models" \
-table_names="user,user_mail"
