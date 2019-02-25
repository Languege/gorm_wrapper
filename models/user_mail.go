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
	UserMailCacheEnable bool = true
	UserMailCacheTimeOut	time.Duration = time.Duration(1) * time.Hour
)

type UserMail struct { 
	UmId  int `gorm:"column:um_id;type:int(11);primary_key;AUTO_INCREMENT"` //用户邮件ID
	Title  string `gorm:"column:title;type:varchar(32);size:32"` //标题
	Content  string `gorm:"column:content;type:text"` //内容
	CreatedAt  int `gorm:"column:created_at;type:int(11)"` //创建时间
	FromUid  int64 `gorm:"column:from_uid;type:bigint(20);default:0"` //邮件发送人
	FromNickname  string `gorm:"column:from_nickname;type:varchar(32);size:32"` //发件人昵称
	UserId  int64 `gorm:"column:user_id;type:bigint(20)"` //邮件收件人
	Status  int8 `gorm:"column:status;type:tinyint(3) unsigned;default:0"` //状态 0-正常 1-已读 2-已删除
	IsReceived  int8 `gorm:"column:is_received;type:tinyint(4);default:0"` //是否领取
	Rewards  string `gorm:"column:rewards;type:text"` //奖励json串
}

type UserMailPaginator struct {
	Data 		[]*UserMail
	CurPage		uint32
	TotalPage 	uint32
	PageSize 	uint32
	TotalSize	uint32
}

func(self UserMail) TableName() string{
	return "user_mail"
}

func(self *UserMail) Insert() error {

	if UserMailCacheEnable {
		err := gorm_wrapper.DB.Create(self).Error
		if err == nil {//写缓存
			self.SyncUserMailCache()
		}

		return err
	}else{
		return gorm_wrapper.DB.Create(self).Error
	}
}

func(self *UserMail) Update(where map[string]interface{}) (err error){
	if len(where) > 0 {//仅更新指定字段
		err = gorm_wrapper.DB.Model(self).Updates(where).Error
	}else{//更新所有字段
		err = gorm_wrapper.DB.Save(self).Error
	}

	if err == nil && UserMailCacheEnable {//更新缓存
		self.SyncUserMailCache()
	}

	return
}

//根据主键查找
func(self *UserMail) FindByPK(UmId int) (err error){

	if UserMailCacheEnable {
		key := self.TableName() + ":"+ "UmId"+strconv.Itoa(int(UmId))

		value, _ := redis_wrapper.Get(key)
		if value != nil {
			err = json.Unmarshal(value, self)
			if err == nil {
				return
			}
		}
	}

	//缓存命中失败
	err = gorm_wrapper.DB.Where(&UserMail{ UmId:UmId }).First(self).Error

	if err == nil && UserMailCacheEnable {
		self.SyncUserMailCache()
	}

	return
}

func(self *UserMail) SyncUserMailCache() {
	key := self.TableName() + ":"+ "UmId"+strconv.Itoa(int(self.UmId))
	bytes, err := json.Marshal(self)
	if err != nil {
		return
	}

	redis_wrapper.Set(key, bytes, int(UserMailCacheTimeOut.Seconds()), 0, false, false)
}

func(self *UserMail) FindOne(where map[string]interface{}) error{
	return gorm_wrapper.DB.Where(where).First(self).Error
}

func(self *UserMail) Query(where map[string]interface{}, limit int32, order map[string]bool) (models []*UserMail, err error){
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

func(self *UserMail) Pagination(where map[string]interface{}, page uint32, pageSize uint32) (*UserMailPaginator, error) {
	offset := (page - 1) * pageSize
	limit := pageSize

	paginator := &UserMailPaginator{Data:[]*UserMail{}, CurPage:page, PageSize:pageSize}

	//查询数据总览
	var err error
	err = gorm_wrapper.DB.Table(self.TableName()).Where(where).Count(&paginator.TotalSize).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination UserMail")
	}

	if paginator.TotalSize <= 0 {
		paginator.TotalPage = 1
		return paginator, nil
	}

	paginator.TotalPage = paginator.TotalSize / paginator.PageSize

	err = gorm_wrapper.DB.Limit(limit).Offset(offset).Where(where).Find(&paginator.Data).Error
	if err != nil {
		return nil, errors.Wrap(err, "Query Error Pagination UserMail")
	}

	return paginator, err
}
