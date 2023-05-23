package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"kayak-backend/global"
	"kayak-backend/model"
	"kayak-backend/utils"
	"net/http"
	"time"
)

type GroupFilter struct {
	ID      *int `json:"id" form:"id"`
	UserId  *int `json:"user_id" form:"user_id"`
	OwnerId *int `json:"owner_id" form:"owner_id"`
	AreaId  *int `json:"area_id" form:"area_id"`
}
type GroupResponse struct {
	Id          int              `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Invitation  string           `json:"invitation"`
	UserId      int              `json:"owner_id"`
	UserInfo    UserInfoResponse `json:"user_info"`
	MemberCount int              `json:"member_count"`
	CreatedAt   time.Time        `json:"created_at"`
	AreaId      int              `json:"area_id"`
	AvatarURL   string           `json:"avatar_url"`
}
type GroupCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AreaId      *int   `json:"area_id"`
}
type AllGroupResponse struct {
	TotalCount int             `json:"total_count"`
	Group      []GroupResponse `json:"group"`
}

// GetGroups godoc
// @Schemes http
// @Description 获取符合filter要求的小组列表
// @Tags Group
// @Param filter query GroupFilter false "筛选条件"
// @Success 200 {object} AllGroupResponse "小组列表"
// @Failure 400 {string} string "请求解析失败"
// @Failure default {string} string "服务器错误"
// @Router /group/all [get]
// @Security ApiKeyAuth
func GetGroups(c *gin.Context) {
	var filter GroupFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.String(http.StatusBadRequest, "请求解析失败")
		return
	}
	var groups []model.Group
	sqlString := `SELECT * FROM "group" WHERE 1 = $1`
	if filter.ID != nil {
		sqlString += fmt.Sprintf(" AND id = %d", *filter.ID)
	}
	if filter.UserId != nil {
		sqlString += fmt.Sprintf(" AND id IN (SELECT group_id FROM group_member WHERE user_id = %d)", *filter.UserId)
	}
	if filter.OwnerId != nil {
		sqlString += fmt.Sprintf(` AND user_id = %d`, *filter.OwnerId)
	}
	if filter.AreaId != nil {
		sqlString += fmt.Sprintf(` AND area_id = %d`, *filter.AreaId)
	}
	if err := global.Database.Select(&groups, sqlString, 1); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	var groupResponses []GroupResponse
	for _, group := range groups {
		user := model.User{}
		sqlString = `SELECT name, email, phone, avatar_url, created_at, nick_name FROM "user" WHERE id = $1`
		if err := global.Database.Get(&user, sqlString, group.UserId); err != nil {
			c.String(http.StatusInternalServerError, "服务器错误")
			return
		}
		userInfo := UserInfoResponse{
			UserId:     user.ID,
			AvatarPath: user.AvatarURL,
			NickName:   user.NickName,
		}
		var count int
		sqlString = `SELECT count(*) FROM group_member WHERE group_id = $1`
		if err := global.Database.Get(&count, sqlString, group.Id); err != nil {
			c.String(http.StatusInternalServerError, "服务器错误")
			return
		}
		groupResponses = append(groupResponses, GroupResponse{
			Id:          group.Id,
			Name:        group.Name,
			Description: group.Description,
			UserId:      group.UserId,
			UserInfo:    userInfo,
			MemberCount: count,
			CreatedAt:   group.CreatedAt,
			AreaId:      group.AreaId,
			AvatarURL:   group.AvatarURL,
		})
	}
	c.JSON(http.StatusOK, AllGroupResponse{
		TotalCount: len(groupResponses),
		Group:      groupResponses,
	})
}

// CreateGroup godoc
// @Schemes http
// @Description 创建小组
// @Tags Group
// @Param group body GroupCreateRequest true "小组信息"
// @Success 200 {object} GroupResponse "小组信息"
// @Failure 400 {string} string "请求解析失败"
// @Failure default {string} string "服务器错误"
// @Router /group/create [post]
// @Security ApiKeyAuth
func CreateGroup(c *gin.Context) {
	var request GroupCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "请求解析失败")
		return
	}
	if request.AreaId == nil {
		request.AreaId = new(int)
		*request.AreaId = 100
	}
	sqlString := `INSERT INTO "group" (name, description, invitation, user_id, created_at, area_id) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var groupId int
	if err := global.Database.Get(&groupId, sqlString, request.Name, request.Description,
		utils.GenerateInvitationCode(4), c.GetInt("UserId"), time.Now().Local(), request.AreaId); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	sqlString = `INSERT INTO group_member (group_id, user_id, created_at, is_admin, is_owner) VALUES ($1, $2, $3, true, true)`
	if _, err := global.Database.Exec(sqlString, groupId, c.GetInt("UserId"), time.Now().Local()); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	var group model.Group
	sqlString = `SELECT * FROM "group" WHERE id = $1`
	if err := global.Database.Get(&group, sqlString, groupId); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.JSON(http.StatusOK, GroupResponse{
		Id:          group.Id,
		Name:        group.Name,
		Description: group.Description,
		Invitation:  group.Invitation,
		UserId:      group.UserId,
		CreatedAt:   group.CreatedAt,
		AreaId:      group.AreaId,
		AvatarURL:   group.AvatarURL,
	})
}

// GetGroupInvitation godoc
// @Schemes http
// @Description 获取小组邀请码
// @Tags Group
// @Param id path int true "小组id"
// @Success 200 {string} string "邀请码"
// @Failure 403 {string} string "没有权限"
// @Failure 404 {string} string "小组不存在
// @Failure default {string} string "服务器错误"
// @Router /group/invitation/{id} [get]
// @Security ApiKeyAuth
func GetGroupInvitation(c *gin.Context) {
	var group model.Group
	sqlString := `SELECT * FROM "group" WHERE id = $1`
	if err := global.Database.Get(&group, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	sqlString = `SELECT group_id FROM group_member WHERE group_id = $1 AND user_id = $2`
	var groupId int
	if err := global.Database.Get(&groupId, sqlString, c.Param("id"), c.GetInt("UserId")); err != nil {
		if role, _ := c.Get("Role"); role != global.ADMIN {
			c.String(http.StatusForbidden, "没有权限")
			return
		}
	}
	c.String(http.StatusOK, group.Invitation)
}

// DeleteGroup godoc
// @Schemes http
// @Description 删除小组
// @Tags Group
// @Param id path int true "小组ID"
// @Success 200 {string} string "删除成功"
// @Failure 403 {string} string "没有权限"
// @Failure 404 {string} string "小组不存在"
// @Failure default {string} string "服务器错误"
// @Router /group/delete/{id} [delete]
// @Security ApiKeyAuth
func DeleteGroup(c *gin.Context) {
	sqlString := `SELECT user_id FROM "group" WHERE id = $1`
	var groupUserId int
	if err := global.Database.Get(&groupUserId, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	if role, _ := c.Get("Role"); groupUserId != c.GetInt("UserId") && role != global.ADMIN {
		c.String(http.StatusForbidden, "没有权限")
		return
	}
	tx := global.Database.MustBegin()
	// 删除小组成员关系
	sqlString = `DELETE FROM group_member WHERE group_id = $1`
	if _, err := tx.Exec(sqlString, c.Param("id")); err != nil {
		if err := tx.Rollback(); err != nil {
			c.String(http.StatusInternalServerError, "服务器错误")
			return
		}
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	sqlString = `DELETE FROM "group" WHERE id = $1`
	if _, err := tx.Exec(sqlString, c.Param("id")); err != nil {
		if err := tx.Rollback(); err != nil {
			c.String(http.StatusInternalServerError, "服务器错误")
			return
		}
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	if err := tx.Commit(); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.String(http.StatusOK, "删除成功")
}

type AllUserResponse struct {
	TotalCount int                `json:"total_count"`
	User       []UserInfoResponse `json:"user"`
}

// GetUsersInGroup godoc
// @Schemes http
// @Description 获取小组成员
// @Tags Group
// @Param id path int true "小组ID"
// @Success 200 {object} AllUserResponse "用户信息"
// @Failure 404 {string} string "小组不存在"
// @Failure default {string} string "服务器错误"
// @Router /group/all_user/{id} [get]
// @Security ApiKeyAuth
func GetUsersInGroup(c *gin.Context) {
	sqlString := `SELECT user_id FROM "group" WHERE id = $1`
	var groupUserId int
	if err := global.Database.Get(&groupUserId, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	var users []model.User
	sqlString = `SELECT * FROM "user" WHERE id IN (SELECT user_id FROM group_member WHERE group_id = $1)`
	if err := global.Database.Select(&users, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	var userResponses []UserInfoResponse
	for _, user := range users {
		userResponses = append(userResponses, UserInfoResponse{
			UserId:     user.ID,
			UserName:   user.Name,
			Email:      user.Email,
			Phone:      user.Phone,
			AvatarPath: user.AvatarURL,
			CreateAt:   user.CreatedAt,
			NickName:   user.NickName,
		})
	}
	c.JSON(http.StatusOK, AllUserResponse{
		TotalCount: len(userResponses),
		User:       userResponses,
	})
}

// AddUserToGroup godoc
// @Schemes http
// @Description 添加用户到小组
// @Tags Group
// @Param id path int true "小组ID"
// @Param user_id query int true "用户ID"
// @Param invitation query string true "邀请码"
// @Success 200 {string} string "添加成功"
// @Failure 403 {string} string "没有权限"
// @Failure 404 {string} string "小组不存在"/"用户不存在"
// @Failure default {string} string "服务器错误"
// @Router /group/add/{id} [post]
// @Security ApiKeyAuth
func AddUserToGroup(c *gin.Context) {
	sqlString := `SELECT invitation FROM "group" WHERE id = $1`
	var invitation string
	if err := global.Database.Get(&invitation, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	sqlString = `SELECT id FROM "user" WHERE id = $1`
	var userId int
	if err := global.Database.Get(&userId, sqlString, c.Query("user_id")); err != nil {
		c.String(http.StatusNotFound, "用户不存在")
		return
	}
	if role, _ := c.Get("Role"); invitation != c.Query("invitation") && role != global.ADMIN {
		c.String(http.StatusForbidden, "没有权限")
		return
	}
	sqlString = `INSERT INTO group_member (user_id, group_id, created_at) VALUES ($1, $2, $3)`
	if _, err := global.Database.Exec(sqlString, c.Query("user_id"), c.Param("id"), time.Now().Local()); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.String(http.StatusOK, "添加成功")
}

// RemoveUserFromGroup godoc
// @Schemes http
// @Description 从小组移除用户
// @Tags Group
// @Param id path int true "小组ID"
// @Param user_id query int true "用户ID"
// @Success 200 {string} string "移除成功"
// @Failure 403 {string} string "没有权限"
// @Failure 404 {string} string "小组不存在"
// @Failure default {string} string "服务器错误"
// @Router /group/remove/{id} [delete]
// @Security ApiKeyAuth
func RemoveUserFromGroup(c *gin.Context) {
	sqlString := `SELECT user_id FROM "group" WHERE id = $1`
	var groupUserId int
	if err := global.Database.Get(&groupUserId, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	if role, _ := c.Get("Role"); groupUserId != c.GetInt("UserId") && role != global.ADMIN {
		c.String(http.StatusForbidden, "没有权限")
		return
	}
	sqlString = `DELETE FROM group_member WHERE user_id = $1 AND group_id = $2`
	if _, err := global.Database.Exec(sqlString, c.Query("user_id"), c.Param("id")); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.String(http.StatusOK, "移除成功")
}

// QuitGroup godoc
// @Schemes http
// @Description 退出小组
// @Tags Group
// @Param id path int true "小组ID"
// @Success 200 {string} string "退出成功"
// @Failure 404 {string} string "小组不存在或用户未加入此小组"
// @Failure 403 {string} string "创建者不能退出创建的小组"
// @Failure default {string} string "服务器错误"
// @Router /group/quit/{id} [delete]
// @Security ApiKeyAuth
func QuitGroup(c *gin.Context) {
	userId := c.GetInt("UserId")
	groupId := c.Param("id")
	sqlString := `SELECT count(*) FROM "group_member" WHERE user_id = $1 AND group_id = $2`
	var count int
	if err := global.Database.Get(&count, sqlString, userId, groupId); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	if count == 0 {
		c.String(http.StatusNotFound, "小组不存在或用户未加入此小组")
		return
	}
	// 如果是创建者，自己不能退出
	sqlString = `SELECT user_id FROM "group" WHERE id = $1`
	var groupUserId int
	if err := global.Database.Get(&groupUserId, sqlString, groupId); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	if groupUserId == userId {
		c.String(http.StatusForbidden, "创建者不能退出创建的小组")
		return
	}
	sqlString = `DELETE FROM group_member WHERE user_id = $1 AND group_id = $2`
	if _, err := global.Database.Exec(sqlString, userId, groupId); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.String(http.StatusOK, "退出成功")
}

type UpdateGroupInfoRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Invitation  *string `json:"invitation"`
	AreaId      *int    `json:"area_id"`
}

// UpdateGroupInfo godoc
// @Schemes http
// @Description 编辑小组信息
// @Tags Group
// @Param id path int true "小组ID"
// @Param group body UpdateGroupInfoRequest true "编辑信息，如果希望不改变的字段传入原值"
// @Success 200 {string} string "编辑成功"
// @Failure 403 {string} string "没有权限"
// @Failure 404 {string} string "小组不存在"
// @Failure default {string} string "服务器错误"
// @Router /group/update/{id} [put]
// @Security ApiKeyAuth
func UpdateGroupInfo(c *gin.Context) {
	var group model.Group
	sqlString := `SELECT * FROM "group" WHERE id = $1`
	if err := global.Database.Get(&group, sqlString, c.Param("id")); err != nil {
		c.String(http.StatusNotFound, "小组不存在")
		return
	}
	if role, _ := c.Get("Role"); group.UserId != c.GetInt("UserId") && role != global.ADMIN {
		c.String(http.StatusForbidden, "没有权限")
		return
	}
	var request UpdateGroupInfoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.String(http.StatusBadRequest, "参数错误")
		return
	}
	if request.Name == nil {
		request.Name = &group.Name
	}
	if request.Description == nil {
		request.Description = &group.Description
	}
	if request.Invitation == nil {
		request.Invitation = &group.Invitation
	}
	if request.AreaId == nil {
		request.AreaId = &group.AreaId
	}
	sqlString = `UPDATE "group" SET name = $1, description = $2, invitation = $3, area_id = $4 WHERE id = $5`
	if _, err := global.Database.Exec(sqlString, request.Name, request.Description,
		request.Invitation, request.AreaId, c.Param("id")); err != nil {
		c.String(http.StatusInternalServerError, "服务器错误")
		return
	}
	c.String(http.StatusOK, "编辑成功")
}
