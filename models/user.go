package models
import (
	"github.com/pkg/errors"
	"Languege/gorm_wrapper"
	"Languege/redis_wrapper"
	"strconv"
	"time"
	"encoding/json"
)

var(
	UserCacheEnable bool = true
	UserCacheTimeOut	time.Duration = time.Duration(1) * time.Hour
)

type User struct { 
	UserId  int64 `gorm:"column:user_id;type:bigint(20);primary_key"` //
	Nickname  string `gorm:"column:nickname;type:varchar(255);size:255"` //昵称
	Gender  int8 `gorm:"column:gender;type:tinyint(3) unsigned;default:0"` //0-未设置 1-男 2-女
	Birthday  int `gorm:"column:birthday;type:int(10) unsigned;default:0"` //出生时间时间戳
	Signature  string `gorm:"column:signature;type:varchar(255);size:255"` //签名
	Avatar  string `gorm:"column:avatar;type:varchar(255);size:255"` //默认头像
	RegisterTime  int `gorm:"column:register_time;type:int(10) unsigned;default:0"` //注册时间时间戳
	LastLogin  int `gorm:"column:last_login;type:int(10) unsigned;default:0"` //上次登录时间时间戳
	LastIp  string `gorm:"column:last_ip;type:varchar(32);size:32"` //用户最近登录ip地址
}

type UserPaginator struct {
	Data 		[]*User
	CurPage		uint32
	TotalPage 	uint32
	PageSize 	uint32
	TotalSize	uint32
}

func(self User) TableName() string{
	return "user"
}

func(self *User) Insert() error {

	if UserCacheEnable {
		err := gorm_wrapper.DB.Create(self).Error
		if err == nil {//写缓存
			self.SyncUserCache()
		}

		return err
	}else{
		return gorm_wrapper.DB.Create(self).Error
	}
}

func(self *User) Update(where map[string]interface{}) (err error){
	if len(where) > 0 {//仅更新指定字段
		err = gorm_wrapper.DB.Model(self).Updates(where).Error
	}else{//更新所有字段
		err = gorm_wrapper.DB.Save(self).Error
	}

	if err == nil && UserCacheEnable {//更新缓存
		self.SyncUserCache()
	}

	return
}

//根据主键查找
func(self *User) FindByPK(UserId int64) (err error){

	if UserCacheEnable {
		key := self.TableName() + ":"+ "UserId"+strconv.Itoa(int(UserId))

		value, _ := redis_wrapper.Get(key)
		if value != nil {
			err = json.Unmarshal(value, self)
			if err == nil {
				return
			}
		}
	}

	//缓存命中失败
	err = gorm_wrapper.DB.Where(&User{ UserId:UserId }).First(self).Error

	if err == nil && UserCacheEnable {
		self.SyncUserCache()
	}

	return
}

func(self *User) SyncUserCache() {
	key := self.TableName() + ":"+ "UserId"+strconv.Itoa(int(self.UserId))
	bytes, err := json.Marshal(self)
	if err != nil {
		return
	}

	redis_wrapper.Set(key, bytes, int(UserCacheTimeOut.Seconds()), 0, false, false)
}

func(self *User) FindOne(where map[string]interface{}) error{
	return gorm_wrapper.DB.Where(where).First(self).Error
}

func(self *User) Query(where map[string]interface{}, limit int32, order map[string]bool) (models []*User, err error){
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

func(self *User) Pagination(where map[string]interface{}, page uint32, pageSize uint32) (*UserPaginator, error) {
	offset := (page - 1) * pageSize
	limit := pageSize

	paginator := &UserPaginator{Data:[]*User{}, CurPage:page, PageSize:pageSize}

	//查询数据总览
	var err error
	err = gorm_wrapper.DB.Table(self.TableName()).Where(where).Count(&paginator.TotalSize).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination User")
	}

	if paginator.TotalSize <= 0 {
		paginator.TotalPage = 1
		return paginator, nil
	}

	paginator.TotalPage = paginator.TotalSize / paginator.PageSize

	err = gorm_wrapper.DB.Limit(limit).Offset(offset).Where(where).Find(&paginator.Data).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination User")
	}

	return paginator, err
}
