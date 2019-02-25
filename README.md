# gorm_wrapper
gorm wrapper

自动生成针对gorm的models文件

## Usage
```bash
go run gen/main.go  -dsn="root:root@tcp(127.0.0.1:3306)/shokudo" \
-db_name="shokudo" \
-out_dir="models"  \
-package="models" \
-table_names="user,user_mail"   #多个表逗号分隔

```

dsn替换成自己的mysql连接
##  测试
修改models_test后测试
```bash
bash models_test.sh

```