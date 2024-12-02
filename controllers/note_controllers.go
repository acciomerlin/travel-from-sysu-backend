package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
)

// UpdateNoteRequest 更新笔记的请求参数
type UpdateNoteRequest struct {
	NoteID      uint   `json:"note_id" binding:"required"` // 笔记 ID
	NoteTitle   string `json:"note_title"`                 // 笔记标题
	NoteContent string `json:"note_content"`               // 笔记内容
	NoteTagList string `json:"note_tag_list"`              // 笔记标签列表
	NoteType    string `json:"note_type"`                  // 笔记类型
	NoteURLs    string `json:"note_URLs"`                  // 笔记相关 URL
}

// UpdateNoteResponse 更新笔记的响应
type UpdateNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// PublishNoteRequest 发布笔记请求参数
type PublishNoteRequest struct {
	NoteTitle     string   `json:"noteTitle" binding:"required"`
	NoteContent   string   `json:"noteContent" binding:"required"`
	NoteCount     int      `json:"noteCount"`
	NoteTagList   []string `json:"noteTagList"` // 使用数组类型
	NoteType      string   `json:"noteType"`
	NoteURLs      string   `json:"noteURLs"`
	NoteCreatorID uint     `json:"noteCreatorID"`
}

// PublishNoteResponse 笔记发布成功的返回信息
type PublishNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// DeleteNoteRequest 删除笔记请求参数
type DeleteNoteRequest struct {
	NoteID uint `json:"note_id" binding:"required"` // 笔记的唯一标识符
}

// DeleteNoteResponse 删除笔记的返回信息
type DeleteNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// PublishNote 发布笔记
func PublishNote(ctx *gin.Context) {
	var req PublishNoteRequest

	// 绑定 JSON 数据到 PublishNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 创建 Note 实例
	note := models.Note{
		NoteTitle:      req.NoteTitle,
		NoteContent:    req.NoteContent,
		ViewCount:      req.NoteCount,
		NoteTagList:    strings.Join(req.NoteTagList, ","), // 将数组转换为字符串存储
		NoteType:       req.NoteType,
		NoteURLs:       req.NoteURLs,
		NoteCreatorID:  req.NoteCreatorID,
		NoteUpdateTime: time.Now().Unix(), // 设置时间戳
	}

	// 保存 Note 到数据库
	if err := global.Db.Create(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记发布失败",
		})
		return
	}

	// 处理 Tag 和 TagNoteRelation
	for _, tagName := range req.NoteTagList {
		var tag models.Tag
		// 检查 Tag 是否已存在
		if err := global.Db.Where("t_name = ?", tagName).First(&tag).Error; err != nil {
			// Tag 不存在，创建新的 Tag
			tag = models.Tag{
				ID:         strconv.FormatInt(time.Now().UnixNano(), 10), // 使用时间戳作为 Tag ID
				TName:      tagName,
				Creator:    strconv.Itoa(int(req.NoteCreatorID)),
				CreateDate: time.Now(),
				UpdateDate: time.Now(),
				UseCount:   1,
			}
			global.Db.Create(&tag)
		} else {
			// Tag 存在，更新 UseCount 和 UpdateTime
			tag.UseCount++
			tag.UpdateDate = time.Now()
			global.Db.Save(&tag)
		}

		// 为每个 Tag 创建对应的 TagNoteRelation 记录
		tagNoteRelation := models.TagNoteRelation{
			NID:        note.NoteID,
			TID:        tag.ID,
			CreatorID:  req.NoteCreatorID,
			CreateDate: time.Now(),
		}
		global.Db.Create(&tagNoteRelation)
	}

	// 成功响应
	ctx.JSON(http.StatusOK, PublishNoteResponse{
		Status: "笔记发布成功",
		Code:   200,
	})
}

// DeleteNote 删除笔记接口
func DeleteNote(ctx *gin.Context) {
	var req DeleteNoteRequest

	// 绑定 JSON 数据到 DeleteNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 查找与笔记相关的 TagNoteRelation 记录
	var relations []models.TagNoteRelation
	if err := global.Db.Where("n_id = ?", req.NoteID).Find(&relations).Error; err == nil {
		fmt.Println("[DEBUG] Found TagNoteRelation records:", relations) // 打印找到的 TagNoteRelation 记录

		for _, relation := range relations {
			// 更新 Tag 的 UseCount
			var tag models.Tag
			if err := global.Db.Where("id = ?", relation.TID).First(&tag).Error; err == nil {
				fmt.Printf("[DEBUG] Found Tag: %+v\n", tag) // 打印找到的 Tag 记录

				// 减少 UseCount
				tag.UseCount--
				if tag.UseCount <= 0 {
					fmt.Printf("[DEBUG] Deleting Tag: %+v\n", tag) // 打印即将删除的 Tag

					// 删除与该 Tag 相关的 TagNoteRelation 记录
					if err := global.Db.Where("t_id = ?", tag.ID).Delete(&models.TagNoteRelation{}).Error; err != nil {
						fmt.Printf("[ERROR] Failed to delete TagNoteRelation for Tag ID %s: %s\n", tag.ID, err)
						continue
					} else {
						fmt.Printf("[DEBUG] Successfully deleted TagNoteRelation for Tag ID: %s\n", tag.ID)
					}

					// 删除 Tag 记录
					if err := global.Db.Delete(&tag).Error; err != nil {
						fmt.Printf("[ERROR] Failed to delete Tag %+v: %s\n", tag, err)
					} else {
						fmt.Printf("[DEBUG] Successfully deleted Tag: %+v\n", tag)
					}
				} else {
					fmt.Printf("[DEBUG] Updating Tag: %+v\n", tag) // 打印即将更新的 Tag
					if err := global.Db.Save(&tag).Error; err != nil {
						fmt.Printf("[ERROR] Failed to update Tag %+v: %s\n", tag, err)
					} else {
						fmt.Printf("[DEBUG] Successfully updated Tag: %+v\n", tag)
					}
				}
			} else {
				fmt.Printf("[DEBUG] Failed to find Tag with ID %s: %s\n", relation.TID, err) // 打印未找到的 Tag 信息
			}

			// 删除当前的 TagNoteRelation 记录
			if err := global.Db.Delete(&relation).Error; err != nil {
				fmt.Printf("[DEBUG] Failed to delete TagNoteRelation %+v: %s\n", relation, err) // 打印删除失败的信息
			} else {
				fmt.Printf("[DEBUG] Successfully deleted TagNoteRelation: %+v\n", relation) // 打印成功删除的信息
			}
		}
	} else {
		fmt.Printf("[DEBUG] Failed to find TagNoteRelation for NoteID %d: %s\n", req.NoteID, err) // 打印未找到 TagNoteRelation 的错误信息
	}

	// 删除 Note
	if err := global.Db.Delete(&models.Note{}, req.NoteID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "删除笔记失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, DeleteNoteResponse{
		Status: "笔记删除成功",
		Code:   200,
	})
}

// UpdateNote 更新笔记接口
func UpdateNote(ctx *gin.Context) {
	var req UpdateNoteRequest

	// 绑定 JSON 数据到 UpdateNoteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, UpdateNoteResponse{
			Status: "失败",
			Code:   400,
			Error:  err.Error(),
		})
		return
	}

	// 根据 NoteID 查找笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", req.NoteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, UpdateNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}

	// 更新 Tag 和 TagNoteRelation
	tagList := strings.Split(req.NoteTagList, ",") // 新的 tag 列表

	// 查找旧的 tag 列表（从数据库中获取 Note 的原始 tag 列表）
	var oldTagList []string
	if note.NoteTagList != "" {
		oldTagList = strings.Split(note.NoteTagList, ",")
	}
	// 打印新的 tagList 和旧的 oldTagList
	fmt.Println("[DEBUG] New Tag List:", tagList)    // 打印新的 tagList
	fmt.Println("[DEBUG] Old Tag List:", oldTagList) // 打印旧的 oldTagList

	// 将旧 tag 和新 tag 转为 map，方便比对
	oldTagMap := make(map[string]bool)
	newTagMap := make(map[string]bool)
	for _, tag := range oldTagList {
		oldTagMap[tag] = true
	}
	for _, tag := range tagList {
		newTagMap[tag] = true
	}

	// 处理删除的 tag（存在于旧 tag 列表但不存在于新 tag 列表）
	for _, oldTag := range oldTagList {
		if !newTagMap[oldTag] {
			var tag models.Tag

			// 更新 Tag 记录
			if err := global.Db.Where("t_name = ?", oldTag).First(&tag).Error; err == nil {
				tag.UseCount--
				// 删除对应的 TagNoteRelation 记录
				var tagNoteRelation models.TagNoteRelation
				global.Db.Where("n_id = ? AND t_id = ?", note.NoteID, tag.ID).Delete(&tagNoteRelation)
				if tag.UseCount <= 0 {
					global.Db.Delete(&tag)
				} else {
					global.Db.Save(&tag)
				}
			}

		}
	}

	// 处理新增的 tag（存在于新 tag 列表但不存在于旧 tag 列表）
	for _, newTag := range tagList {
		if !oldTagMap[newTag] {
			var tag models.Tag
			if err := global.Db.Where("t_name = ?", newTag).First(&tag).Error; err != nil {
				tag = models.Tag{
					ID:         strconv.FormatInt(time.Now().UnixNano(), 10),
					TName:      newTag,
					Creator:    strconv.Itoa(int(note.NoteCreatorID)),
					CreateDate: time.Now(),
					UpdateDate: time.Now(),
					UseCount:   1,
				}
				global.Db.Create(&tag)
			} else {
				tag.UseCount++
				tag.UpdateDate = time.Now()
				global.Db.Save(&tag)
			}

			// 为 Tag 创建新的 TagNoteRelation 记录
			tagNoteRelation := models.TagNoteRelation{
				NID:        note.NoteID,
				TID:        tag.ID,
				CreatorID:  note.NoteCreatorID,
				CreateDate: time.Now(),
			}
			global.Db.Create(&tagNoteRelation)
		}
	}

	// 更新笔记内容
	if req.NoteTitle != "" {
		note.NoteTitle = req.NoteTitle
	}
	if req.NoteContent != "" {
		note.NoteContent = req.NoteContent
	}
	if req.NoteTagList != "" {
		note.NoteTagList = req.NoteTagList
	}
	if req.NoteType != "" {
		note.NoteType = req.NoteType
	}
	if req.NoteURLs != "" {
		note.NoteURLs = req.NoteURLs
	}
	note.NoteUpdateTime = time.Now().Unix() // 更新时间戳

	// 保存更新到数据库
	if err := global.Db.Save(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, UpdateNoteResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记更新失败",
		})
		return
	}

	// 成功响应
	ctx.JSON(http.StatusOK, UpdateNoteResponse{
		Status: "笔记更新成功",
		Code:   200,
	})
}

// GetFoNotes 获取用户关注用户的帖子，支持游标分页，使用时间戳
func GetFoNotes(ctx *gin.Context) {
	// 获取请求参数
	userID := ctx.Query("user_id")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（时间戳）

	// 参数校验
	if userID == "" || num == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "参数缺失",
		})
		return
	}

	// 默认最大条数
	limit := 30
	if n, err := strconv.Atoi(num); err == nil && n > 0 && n < 30 {
		limit = n
	}

	// 查询用户关注的用户 ID
	var followers []models.Follower
	if err := global.Db.Where("uid = ?", userID).Find(&followers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询关注信息失败",
		})
		return
	}

	// 获取关注的用户 ID 列表
	var followedUserIDs []uint
	for _, follower := range followers {
		followedUserIDs = append(followedUserIDs, follower.Fid)
	}

	// 构造查询条件
	query := global.Db.Where("note_creator_id IN ?", followedUserIDs)
	if cursor != "" {
		// 游标为时间戳（Unix 时间）
		if timestamp, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			query = query.Where("note_update_time < ?", timestamp) // 返回最近更新的帖子，所以 <
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标参数",
			})
			return
		}
	}

	// 查询帖子数据
	var notes []models.Note
	if err := query.Order("note_update_time DESC").Limit(limit).Find(&notes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询笔记失败",
		})
		return
	}

	// 构造返回结果
	var responseNotes []gin.H
	var nextCursor string
	for _, note := range notes {
		responseNotes = append(responseNotes, gin.H{
			"note_id":          note.NoteID,
			"note_title":       note.NoteTitle,
			"note_content":     note.NoteContent,
			"like_counts":      note.LikeCounts,
			"collect_counts":   note.CollectCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime, // 时间戳直接返回，前端要解析！
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.FormatInt(notes[len(notes)-1].NoteUpdateTime, 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}
