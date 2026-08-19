package model

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type Invitation struct {
	Id          int            `json:"id"`
	UserId      int            `json:"user_id"`
	Key         string         `json:"key" gorm:"type:varchar(64);uniqueIndex"`
	Status      int            `json:"status" gorm:"default:1"` // 1: Enabled, 2: Disabled, 3: Used
	Name        string         `json:"name" gorm:"index"`
	Quota       int            `json:"quota" gorm:"default:100"`
	Group       string         `json:"group" gorm:"type:varchar(32);default:''"`
	CreatedTime int64          `json:"created_time" gorm:"bigint"`
	UsedTime    int64          `json:"used_time" gorm:"bigint"`
	Count       int            `json:"count" gorm:"-:all"` // only for api request
	UsedUserId  int            `json:"used_user_id"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
	ExpiredTime int64          `json:"expired_time" gorm:"bigint"` // 过期时间，0 表示不过期
}

func GetAllInvitations(startIdx int, num int) (invitations []*Invitation, total int64, err error) {
	err = DB.Model(&Invitation{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = DB.Order("id desc").Limit(num).Offset(startIdx).Find(&invitations).Error
	return invitations, total, err
}

func SearchInvitations(keyword string, status string, startIdx int, num int) (invitations []*Invitation, total int64, err error) {
	query := DB.Model(&Invitation{})

	if keyword != "" {
		if id, err := strconv.Atoi(keyword); err == nil {
			query = query.Where("id = ? OR name LIKE ? OR "+commonKeyCol+" LIKE ?", id, keyword+"%", keyword+"%")
		} else {
			query = query.Where("name LIKE ? OR "+commonKeyCol+" LIKE ?", keyword+"%", keyword+"%")
		}
	}

	if status != "" {
		now := common.GetTimestamp()
		switch status {
		case "expired":
			query = query.Where(
				"status = ? AND expired_time != 0 AND expired_time < ?",
				common.InvitationCodeStatusEnabled,
				now,
			)
		case strconv.Itoa(common.InvitationCodeStatusEnabled):
			query = query.Where(
				"status = ? AND (expired_time = 0 OR expired_time >= ?)",
				common.InvitationCodeStatusEnabled,
				now,
			)
		case strconv.Itoa(common.InvitationCodeStatusDisabled):
			query = query.Where("status = ?", common.InvitationCodeStatusDisabled)
		case strconv.Itoa(common.InvitationCodeStatusUsed):
			query = query.Where("status = ?", common.InvitationCodeStatusUsed)
		}
	}

	err = query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("id desc").Limit(num).Offset(startIdx).Find(&invitations).Error
	return invitations, total, err
}

func GetInvitationById(id int) (*Invitation, error) {
	if id == 0 {
		return nil, errors.New("id 为空！")
	}
	invitation := Invitation{Id: id}
	var err error = nil
	err = DB.First(&invitation, "id = ?", id).Error
	return &invitation, err
}

func GetInvitationByKey(key string) (*Invitation, error) {
	if key == "" {
		return nil, errors.New("未提供邀请码")
	}
	invitation := Invitation{}
	err := DB.Where(commonKeyCol+" = ?", key).First(&invitation).Error
	return &invitation, err
}

func (invitation *Invitation) Insert() error {
	return DB.Create(invitation).Error
}

func BatchInsertInvitations(invitations []*Invitation) error {
	if len(invitations) == 0 {
		return nil
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&invitations).Error
	})
}

func (invitation *Invitation) SelectUpdate() error {
	return DB.Model(invitation).Select("used_time", "status", "used_user_id").Updates(invitation).Error
}

func (invitation *Invitation) Update() error {
	return DB.Model(invitation).Select("name", "status", "quota", "group", "expired_time").Updates(invitation).Error
}

func (invitation *Invitation) Delete() error {
	return DB.Delete(invitation).Error
}

func DeleteInvitationById(id int) (err error) {
	if id == 0 {
		return errors.New("id 为空！")
	}
	invitation := Invitation{Id: id}
	err = DB.Where("id = ?", id).First(&invitation).Error
	if err != nil {
		return err
	}
	return invitation.Delete()
}

func BatchDeleteInvitations(ids []int) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := DB.Where("id IN ?", ids).Delete(&Invitation{})
	return result.RowsAffected, result.Error
}

func DeleteInvalidInvitations() (int64, error) {
	now := common.GetTimestamp()
	result := DB.Where(
		"status IN ? OR (status = ? AND expired_time != 0 AND expired_time < ?)",
		[]int{common.InvitationCodeStatusUsed, common.InvitationCodeStatusDisabled},
		common.InvitationCodeStatusEnabled,
		now,
	).Delete(&Invitation{})
	return result.RowsAffected, result.Error
}
