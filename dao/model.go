package dao

import (
	"log"

	"github.com/ah-its-andy/smsresender/db"
	"gorm.io/gorm"
)

func AutoMigrate() {
	gdb, err := db.OpenConnection(db.DefaultOptions())
	if err != nil {
		panic(err)
	}
	// if db.DefaultOptions().DriverType == "mysql" {
	// 	gdb = gdb.Set("gorm:table_options", " ENGINE=InnoDB ")
	// }
	if !gdb.Migrator().HasTable(&SmsModel{}) {
		err = gdb.Migrator().AutoMigrate(&SmsModel{})
		if err != nil {
			panic(err)
		}
	}
}

type SmsModel struct {
	gorm.Model
	Device    string `gorm:"column:device;type:varchar(32);size:32"`
	MessageId string `gorm:"column:message_id;type:varchar(32);size:32"`
	Sender    string `gorm:"column:sender;type:varchar(32);size:32"`
	Content   string `gorm:"column:content;type:text"`
	RecTime   string `gorm:"column:recTime;type:datetime"`
	State     uint32 `gorm:"column:state;type:tinyint;default:0"`
}

func (model *SmsModel) TableName() string {
	return "sms"
}

func GetSmsList(gdb *gorm.DB, states ...uint32) ([]*SmsModel, error) {
	var receivers []*SmsModel
	err := gdb.Table("sms").
		Where("state IN (?)", states).Order("created_at ASC").Find(&receivers).Error
	if err != nil {
		return nil, err
	}
	return receivers, nil
}

func GetSmsById(gdb *gorm.DB, id uint) (*SmsModel, error) {
	var receivers []*SmsModel
	err := gdb.Table("sms").
		Where("id=?", id).Order("created_at ASC").Find(&receivers).Error
	if err != nil {
		return nil, err
	}
	if len(receivers) == 0 {
		return nil, nil
	}
	return receivers[0], nil
}

func FindOneSms(gdb *gorm.DB, device, messageId string) (*SmsModel, error) {
	var receivers []*SmsModel
	err := gdb.Table("sms").
		Where("device=? AND message_id=?", device, messageId).Order("created_at ASC").Find(&receivers).Error
	if err != nil {
		return nil, err
	}
	if len(receivers) == 0 {
		return nil, nil
	}
	return receivers[0], nil
}

func CreateSms(gdb *gorm.DB, entity *SmsModel) error {
	model, err := FindOneSms(gdb, entity.Device, entity.MessageId)
	if err != nil {
		return err
	}
	if model != nil {
		log.Printf("Duplicate entry %s-%s-%s", entity.Device, entity.Sender, entity.Content)
		return nil
	}
	entity.ID = NextID()
	err = gdb.Table(entity.TableName()).Create(entity).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateSms(gdb *gorm.DB, entity *SmsModel) error {
	return gdb.Table(entity.TableName()).Updates(entity).Error
}
