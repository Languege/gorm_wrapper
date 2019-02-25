package main

import (
	"flag"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"os"
	"text/template"
	"sync"
)

var(
	tableNames,outPath,DSN, DBName,Package string
)
var funcMap = template.FuncMap{}

func init(){
	flag.StringVar(&DSN, "dsn", "","-dsn=	#dsn string for connection")
	flag.StringVar(&tableNames, "table_names", "", "-table_names=   #splited by ,")
	flag.StringVar(&outPath, "out_dir", "", "-out_dir=")
	flag.StringVar(&DBName, "db_name", "","-db_name=")
	flag.StringVar(&Package, "package", "","-package=  #package name")


	//FindByPK函数 参数
	funcMap["FindByPKParams"] = func(primaryKey []*Column) string {

		paramsString := ""
		for _, v := range primaryKey {
			paramsString += v.Name + " " + v.Type + ","
		}

		paramsString = strings.TrimRight(paramsString, ",")
		return paramsString
	}

	funcMap["StringKeyGen"] = func(primaryKey []*Column) string {
		keyString := ""
		for _, v := range primaryKey {
			switch v.Type {
			case "string":
				keyString += "\"" + v.Name + "\"+" + v.Name +  "+\"_\"+"
			default:
				keyString += "\"" + v.Name + "\"+strconv.Itoa(int(" + v.Name + "))+\"_\"+"
			}
		}

		keyString = strings.TrimRight(keyString, "+\"_\"+")
		return keyString
	}

	//Self key表达式
	funcMap["SelfKeyGen"] = func(primaryKey []*Column) string {
		keyString := ""
		for _, v := range primaryKey {
			switch v.Type {
			case "string":
				keyString += "\"" + v.Name + "\"+self." + v.Name + "+\"_\"+"
			default:
				keyString += "\"" + v.Name + "\"+strconv.Itoa(int(self." + v.Name + "))+\"_\"+"
			}
		}

		keyString = strings.TrimRight(keyString, "+\"_\"+")
		return keyString
	}

	//Get函数 FindOne查询条件生成
	funcMap["GetQueryConditionGen"] = func(primaryKey []*Column) string {

		keyString := ""
		for _, v := range primaryKey {
			keyString += v.Name + ":" + v.Name + ","
		}

		keyString = strings.TrimRight(keyString, ",")
		return keyString
	}

	//Collection和队列元素数据类型，
	funcMap["CollectionItemType"] = func(primaryKey []*Column) string {
		if len(primaryKey) > 1 {//联合主键
			return "string"
		}else{
			return primaryKey[0].Type
		}
	}

	//PersistentItem函数 FindOne查询条件生成
	funcMap["PersistentItemQueryConditionGen"] = func(primaryKey []*Column) string {

		keyString := ""
		for _, v := range primaryKey {
			keyString += v.Name + ":value." + v.Name + ","
		}

		keyString = strings.TrimRight(keyString, ",")
		return keyString
	}
}

type SchemaColumn struct {
	ColumnName string `gorm:"column:COLUMN_NAME"`
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
	ColumnType string `gorm:"column:COLUMN_TYPE"`
	DataType string	`gorm:"column:DATA_TYPE"`
	ColumnKey string `gorm:"column:COLUMN_KEY"`
	ColumnComment string `gorm:"column:COLUMN_COMMENT"`
	Extra string `gorm:"column:EXTRA"`
	CharMaxLength int64 `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
}

type Column struct {
	Name 	string
	Type 	string
	Tags 	string
	Comment string
}

type TableDesc struct {
	Package 	string
	TableName	string
	StructName 	string
	Columns 	[]*Column
	PrimaryKey 	[]*Column
}

//camel string, xx_yy to XxYy
func camelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

// typeMapping maps SQL data type to corresponding Go data type
var typeMappingMysql = map[string]string{
	"int":                "int", // int signed
	"integer":            "int",
	"tinyint":            "int8",
	"smallint":           "int16",
	"mediumint":          "int32",
	"bigint":             "int64",
	"int unsigned":       "uint", // int unsigned
	"integer unsigned":   "uint",
	"tinyint unsigned":   "uint8",
	"smallint unsigned":  "uint16",
	"mediumint unsigned": "uint32",
	"bigint unsigned":    "uint64",
	"bit":                "uint64",
	"bool":               "bool",   // boolean
	"enum":               "string", // enum
	"set":                "string", // set
	"varchar":            "string", // string & text
	"char":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"text":               "string",
	"longtext":           "string",
	"blob":               "string", // blob
	"tinyblob":           "string",
	"mediumblob":         "string",
	"longblob":           "string",
	"date":               "time.Time", // time
	"datetime":           "time.Time",
	"timestamp":          "time.Time",
	"time":               "time.Time",
	"float":              "float32", // float & decimal
	"double":             "float64",
	"decimal":            "float64",
	"binary":             "string", // binary
	"varbinary":          "string",
	"year":               "int16",
}

const (
	tpl string = `package {{.Package}}
import (
	"github.com/pkg/errors"
	"Languege/gorm_wrapper"
	"Languege/redis_wrapper"
	"strconv"
	"time"
	"encoding/json"
)

var(
	{{.StructName}}CacheEnable bool = true
	{{.StructName}}CacheTimeOut	time.Duration = time.Duration(1) * time.Hour
)

type {{.StructName}} struct { {{range .Columns}}
	{{.Name}}  {{.Type}} {{.Tags}} {{.Comment}}{{end}}
}

type {{.StructName}}Paginator struct {
	Data 		[]*{{.StructName}}
	CurPage		uint32
	TotalPage 	uint32
	PageSize 	uint32
	TotalSize	uint32
}

func(self {{.StructName}}) TableName() string{
	return "{{.TableName}}"
}

func(self *{{.StructName}}) Insert() error {

	if {{.StructName}}CacheEnable {
		err := gorm_wrapper.DB.Create(self).Error
		if err == nil {//写缓存
			self.Sync{{.StructName}}Cache()
		}

		return err
	}else{
		return gorm_wrapper.DB.Create(self).Error
	}
}

func(self *{{.StructName}}) Update(where map[string]interface{}) (err error){
	if len(where) > 0 {//仅更新指定字段
		err = gorm_wrapper.DB.Model(self).Updates(where).Error
	}else{//更新所有字段
		err = gorm_wrapper.DB.Save(self).Error
	}

	if err == nil && {{.StructName}}CacheEnable {//更新缓存
		self.Sync{{.StructName}}Cache()
	}

	return
}

//根据主键查找
func(self *{{.StructName}}) FindByPK({{FindByPKParams .PrimaryKey}}) (err error){

	if {{.StructName}}CacheEnable {
		key := self.TableName() + ":"+ {{StringKeyGen .PrimaryKey}}

		value, _ := redis_wrapper.Get(key)
		if value != nil {
			err = json.Unmarshal(value, self)
			if err == nil {
				return
			}
		}
	}

	//缓存命中失败
	err = gorm_wrapper.DB.Where(&{{.StructName}}{ {{GetQueryConditionGen .PrimaryKey }} }).First(self).Error

	if err == nil && {{.StructName}}CacheEnable {
		self.Sync{{.StructName}}Cache()
	}

	return
}

func(self *{{.StructName}}) Sync{{.StructName}}Cache() {
	key := self.TableName() + ":"+ {{SelfKeyGen .PrimaryKey}}
	bytes, err := json.Marshal(self)
	if err != nil {
		return
	}

	redis_wrapper.Set(key, bytes, int({{.StructName}}CacheTimeOut.Seconds()), 0, false, false)
}

func(self *{{.StructName}}) FindOne(where map[string]interface{}) error{
	return gorm_wrapper.DB.Where(where).First(self).Error
}

func(self *{{.StructName}}) Query(where map[string]interface{}, limit int32, order map[string]bool) (models []*{{.StructName}}, err error){
	db := gorm_wrapper.DB.Table(self.TableName()).Limit(limit).Where(where)
	if order != nil {
		for k, v := range order {
			db = db.Order(k, v)
		}

		err = db.Find(&models).Error
	}else{
		err = db.Find(&models).Error
	}
	return
}

func(self *{{.StructName}}) Pagination(where map[string]interface{}, page uint32, pageSize uint32) (*{{.StructName}}Paginator, error) {
	offset := (page - 1) * pageSize
	limit := pageSize

	paginator := &{{.StructName}}Paginator{Data:[]*{{.StructName}}{}, CurPage:page, PageSize:pageSize}

	//查询数据总览
	var err error
	err = gorm_wrapper.DB.Table(self.TableName()).Where(where).Count(&paginator.TotalSize).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination {{.StructName}}")
	}

	if paginator.TotalSize <= 0 {
		paginator.TotalPage = 1
		return paginator, nil
	}

	paginator.TotalPage = paginator.TotalSize / paginator.PageSize

	err = gorm_wrapper.DB.Limit(limit).Offset(offset).Where(where).Find(&paginator.Data).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination {{.StructName}}")
	}

	return paginator, err
}
`

)

var(
	DB *gorm.DB
	WaitGroup sync.WaitGroup
)


func genOrm(tableName string) {
	defer WaitGroup.Done()
	sql := "SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME='"+tableName+"' AND TABLE_SCHEMA='"+DBName+"'"
	fmt.Println(sql)
	result := &TableDesc{Package:Package,TableName:tableName,StructName:camelString(tableName)}
	data := []*SchemaColumn{}

	err := DB.Raw(sql).Scan(&data).Error

	if err != nil{
		panic(err)
	}



	for _, col := range data {
		fmt.Println("%+v", col)

		//gorm标签处理
		var tags []string
		//列名
		tags = append(tags, fmt.Sprintf("column:%s", col.ColumnName))
		tags = append(tags, fmt.Sprintf("type:%s", col.ColumnType))

		//主键
		if col.ColumnKey == "PRI" {
			tags = append(tags, "primary_key")
		}

		//是否自增
		if col.Extra ==  "auto_increment" {
			tags = append(tags, "AUTO_INCREMENT")
		}

		//varchar 对应的size
		if col.DataType == "varchar" {
			tags = append(tags, fmt.Sprintf("size:%d", col.CharMaxLength))
		}

		if col.ColumnDefault != "" {
			tags = append(tags, fmt.Sprintf("default:%s", col.ColumnDefault))
		}

		column := &Column{}

		column.Tags = fmt.Sprintf("`gorm:\"%s\"`", strings.Join(tags, ";"))
		column.Comment = fmt.Sprintf("//%s", col.ColumnComment)
		column.Name = camelString(col.ColumnName)
		column.Type = typeMappingMysql[col.DataType]

		if col.ColumnKey == "PRI" {
			result.PrimaryKey = append(result.PrimaryKey,  column)
		}

		result.Columns = append(result.Columns, column)
	}

	fmt.Println(result.Columns)

	outFilename := outPath + "/" + tableName + ".go"
	fmt.Println(outFilename)

	outFile, err := os.Create(outFilename)
	if err != nil {
		panic(err.Error())
	}

	tmpl, err := template.New("struct_gen").Funcs(funcMap).Parse(tpl)
	if err != nil { panic(err) }
	err = tmpl.Execute(outFile, result)
	if err != nil { panic(err) }

	outFile.Close()
}

func main(){
	flag.Parse()
	//获取数据表结构
	var err error
	DB, err = gorm.Open("mysql", DSN)
	if err != nil {
		panic(err)
	}

	tableNames := strings.Split(tableNames, ",")

	for _, tableName := range tableNames{
		WaitGroup.Add(1)
		go genOrm(tableName)
	}

	WaitGroup.Wait()
}