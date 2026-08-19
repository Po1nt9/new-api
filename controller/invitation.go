package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

func GetAllInvitations(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	invitations, total, err := model.GetAllInvitations(pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(invitations)
	common.ApiSuccess(c, pageInfo)
}

func SearchInvitations(c *gin.Context) {
	keyword := c.Query("keyword")
	status := c.Query("status")
	pageInfo := common.GetPageQuery(c)
	invitations, total, err := model.SearchInvitations(keyword, status, pageInfo.GetStartIdx(), pageInfo.GetPageSize())
	if err != nil {
		common.ApiError(c, err)
		return
	}
	pageInfo.SetTotal(int(total))
	pageInfo.SetItems(invitations)
	common.ApiSuccess(c, pageInfo)
}

func GetInvitation(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		common.ApiError(c, err)
		return
	}
	invitation, err := model.GetInvitationById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    invitation,
	})
}

type AddInvitationRequest struct {
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Key         string `json:"key"`
	Quota       int    `json:"quota"`
	Group       string `json:"group"`
	Count       int    `json:"count"`
	ExpiredTime int64  `json:"expired_time"`
}

func AddInvitation(c *gin.Context) {
	req := AddInvitationRequest{}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Prefix = strings.TrimSpace(req.Prefix)
	req.Key = strings.TrimSpace(req.Key)
	req.Group = strings.TrimSpace(req.Group)

	if utf8.RuneCountInString(req.Name) == 0 || utf8.RuneCountInString(req.Name) > 20 {
		common.ApiErrorI18n(c, i18n.MsgInvitationNameLength)
		return
	}
	if req.Count <= 0 {
		req.Count = 1
	}
	if req.Count > 100 {
		common.ApiErrorI18n(c, i18n.MsgInvitationCountMax)
		return
	}
	if valid, msg := validateExpiredTime(c, req.ExpiredTime); !valid {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
		return
	}

	var keys []string
	var invitations []*model.Invitation
	userId := c.GetInt("id")
	now := common.GetTimestamp()

	for i := 0; i < req.Count; i++ {
		key := req.Key
		if key == "" || req.Count > 1 {
			if req.Prefix != "" {
				key = fmt.Sprintf("%s%s", req.Prefix, common.GetUUID())
			} else {
				key = fmt.Sprintf("inv_%s", common.GetUUID())
			}
		}
		cleanInvitation := &model.Invitation{
			UserId:      userId,
			Name:        req.Name,
			Key:         key,
			CreatedTime: now,
			Quota:       req.Quota,
			Group:       req.Group,
			Status:      common.InvitationCodeStatusEnabled,
			ExpiredTime: req.ExpiredTime,
		}
		invitations = append(invitations, cleanInvitation)
		keys = append(keys, key)
	}

	err = model.BatchInsertInvitations(invitations)
	if err != nil {
		common.SysError("failed to insert invitations: " + err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": i18n.T(c, i18n.MsgInvitationCreateFailed),
		})
		return
	}

	recordManageAudit(c, "invitation.create", map[string]interface{}{
		"name":  req.Name,
		"count": req.Count,
		"quota": logger.LogQuota(req.Quota),
		"group": req.Group,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
}

func DeleteInvitation(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := model.DeleteInvitationById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "invitation.delete", map[string]interface{}{
		"id": id,
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

type BatchDeleteInvitationRequest struct {
	Ids []int `json:"ids"`
}

func BatchDeleteInvitations(c *gin.Context) {
	var req BatchDeleteInvitationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil || len(req.Ids) == 0 {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}
	rows, err := model.BatchDeleteInvitations(req.Ids)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "invitation.batch_delete", map[string]interface{}{
		"ids":           req.Ids,
		"rows_affected": rows,
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    rows,
	})
}

func UpdateInvitation(c *gin.Context) {
	statusOnly := c.Query("status_only")
	invitation := model.Invitation{}
	err := c.ShouldBindJSON(&invitation)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	cleanInvitation, err := model.GetInvitationById(invitation.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if statusOnly == "" {
		if valid, msg := validateExpiredTime(c, invitation.ExpiredTime); !valid {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": msg})
			return
		}
		cleanInvitation.Name = strings.TrimSpace(invitation.Name)
		cleanInvitation.Quota = invitation.Quota
		cleanInvitation.Group = strings.TrimSpace(invitation.Group)
		cleanInvitation.ExpiredTime = invitation.ExpiredTime
	}
	if statusOnly != "" {
		cleanInvitation.Status = invitation.Status
	}
	err = cleanInvitation.Update()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "invitation.update", map[string]interface{}{
		"id":     cleanInvitation.Id,
		"status": cleanInvitation.Status,
		"quota":  cleanInvitation.Quota,
		"group":  cleanInvitation.Group,
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    cleanInvitation,
	})
}

func DeleteInvalidInvitation(c *gin.Context) {
	rows, err := model.DeleteInvalidInvitations()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	recordManageAudit(c, "invitation.delete_invalid", map[string]interface{}{
		"rows_affected": rows,
	})
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    rows,
	})
}
