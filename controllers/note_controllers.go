package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"travel-from-sysu-backend/global"
	"travel-from-sysu-backend/models"
	"travel-from-sysu-backend/oss"
)

type NewNote struct {
	NoteID           uint     `gorm:"primaryKey;autoIncrement;autoIncrementStart:100001" json:"note_id"` // 主键 ID
	NoteTitle        string   `json:"note_title"`                                                        // 笔记标题
	NoteContent      string   `json:"note_content"`                                                      // 笔记内容
	ViewCount        int      `json:"view_count"`                                                        // 浏览计数
	NoteTagList      string   `json:"note_tag_list"`                                                     // 笔记标签列表（字符串类型）
	NoteType         string   `json:"note_type"`
	NoteURLS         []string `json:"note_urls"`
	NoteCreatorID    uint     `gorm:"not null;index" json:"note_creator_id"` // 创建者 ID（外键）
	NoteUpdateTime   int64    `json:"note_update_time"`                      // 笔记更新时间 (Unix 时间戳)
	LikeCounts       int      `json:"like_counts"`
	CollectCounts    int      `json:"collect_counts"`
	CommentCounts    int      `json:"comment_counts"`
	IsFindingBuddy   int      `json:"is_finding_buddy"`  // 是否是找旅伴帖子 (0: 否, 1: 是)
	BuddyDescription string   `json:"buddy_description"` // 找旅伴的需求描述
}

// GetNotesByCreatorIDResponse 获取笔记响应结构
type GetNotesByCreatorIDResponse struct {
	Status     string    `json:"status"`
	Code       int       `json:"code"`
	Notes      []NewNote `json:"notes,omitempty"`
	NextCursor string    `json:"next_cursor"`
	Error      string    `json:"error,omitempty"`
}

// GetNoteResponse 获取笔记的响应结构
type GetNoteResponse struct {
	Status           string   `json:"status"`                                                            // 响应状态
	Code             int      `json:"code"`                                                              // 响应代码
	NoteID           uint     `gorm:"primaryKey;autoIncrement;autoIncrementStart:100001" json:"note_id"` // 主键 ID
	NoteTitle        string   `json:"note_title"`                                                        // 笔记标题
	NoteContent      string   `json:"note_content"`                                                      // 笔记内容
	ViewCount        uint     `json:"view_count"`                                                        // 浏览计数
	NoteTagList      string   `json:"note_tag_list"`                                                     // 笔记标签列表（字符串类型）
	NoteType         string   `json:"note_type"`                                                         // 笔记类型 	// 笔记相关 URL
	NoteCreatorID    uint     `gorm:"not null;index" json:"note_creator_id"`                             // 创建者 ID（外键）
	NoteUpdateTime   int64    `json:"note_update_time"`                                                  // 笔记更新时间 (Unix 时间戳)
	LikeCounts       uint     `json:"like_counts"`
	CollectCounts    uint     `json:"collect_counts"`
	CommentCounts    uint     `json:"comment_counts"`
	IsFindingBuddy   int      `json:"is_finding_buddy"`  // 是否是找旅伴帖子 (0: 否, 1: 是)
	BuddyDescription string   `json:"buddy_description"` // 找旅伴的需求描述
	Error            string   `json:"error,omitempty"`   // 错误信息
	NoteURLS         []string `json:"note_urls"`
}

// UpdateNoteRequest 更新笔记的请求参数
type UpdateNoteRequest struct {
	NoteID           uint   `json:"note_id" binding:"required"` // 笔记 ID
	NoteTitle        string `json:"note_title"`                 // 笔记标题
	NoteContent      string `json:"note_content"`               // 笔记内容
	NoteTagList      string `json:"note_tag_list"`              // 笔记标签列表
	NoteType         string `json:"note_type"`                  // 笔记类型
	NoteURLs         string `json:"note_URLs"`                  // 笔记相关 URL
	IsFindingBuddy   int    `json:"is_finding_buddy"`           // 是否是找旅伴帖子 (0: 否, 1: 是)
	BuddyDescription string `json:"buddy_description"`          // 找旅伴的需求描述
}

// UpdateNoteResponse 更新笔记的响应
type UpdateNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
}

// PublishNoteRequest 发布笔记请求参数
type PublishNoteRequest struct {
	NoteTitle      string   `json:"noteTitle" binding:"required"`
	NoteContent    string   `json:"noteContent" binding:"required"`
	NoteCount      int      `json:"noteCount"`
	NoteTagList    []string `json:"noteTagList"` // 使用数组类型
	NoteType       string   `json:"noteType"`
	NoteURLs       string   `json:"noteURLs"`
	NoteCreatorID  uint     `json:"noteCreatorID"`
	IsFindingBuddy int      `json:"is_finding_buddy"` // 是否是找旅伴帖子 (0: 否, 1: 是)
}

// PublishNoteResponse 笔记发布成功的返回信息
type PublishNoteResponse struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Error  string `json:"error,omitempty"`
	NID    uint   `json:"nid"`
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

// cleanupUploadedFiles 删除已上传的文件
func cleanupUploadedFiles(urls []string) {
	for _, url := range urls {
		if err := oss.DeleteFileFromAliyunOss(url); err != nil {
			log.Printf("删除文件失败: %v", err)
		}
	}
}

// PublishNoteWithPics 发布笔多图笔记（含上传笔记图片至oss）
func PublishNoteWithPics(ctx *gin.Context) {
	// 接收笔记信息
	noteTitle := ctx.PostForm("note_title")
	noteContent := ctx.PostForm("note_content")
	noteTagList := ctx.PostForm("note_tag_list")
	noteType := ctx.PostForm("note_type")
	noteCreatorID := ctx.PostForm("note_creator_id")
	isFindingBuddy := ctx.PostForm("is_finding_buddy")
	buddyDescription := ctx.PostForm("buddy_description")

	if noteTitle == "" || noteContent == "" || noteCreatorID == "" || isFindingBuddy != "0" || isFindingBuddy != "1" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要参数",
		})
		return
	}

	// 如果是找旅伴帖子，检查 buddy_description 是否为空
	if isFindingBuddy == "1" && buddyDescription == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "找旅伴帖子必须提供需求描述",
		})
		return
	}

	creatorID, err := strconv.Atoi(noteCreatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "创建者ID格式错误",
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := global.Db.First(&user, "user_id = ?", creatorID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "用户不存在",
		})
		return
	}

	// 处理文件
	files := ctx.Request.MultipartForm.File["files"]
	if files == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  "没有上传文件，或者请求格式不正确",
		})
		return
	}

	var uploadedURLs []string

	// 上传文件到 OSS
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			// 删除已上传的文件
			cleanupUploadedFiles(uploadedURLs)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "失败",
				"code":   400,
				"error":  "不支持的文件类型",
			})
			return
		}

		url, err := oss.UploadFileToAliyunOss(file, "note_pics")
		if err != nil {
			// 删除已上传的文件
			cleanupUploadedFiles(uploadedURLs)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "图片上传失败",
			})
			return
		}
		uploadedURLs = append(uploadedURLs, url)
	}

	// 转换 URL 为 JSON
	noteURLsJSON, err := json.Marshal(uploadedURLs)
	if err != nil {
		// 删除已上传的文件
		cleanupUploadedFiles(uploadedURLs)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "图片URL序列化失败",
		})
		return
	}

	// 创建 Note 记录
	isFindingBuddyInt, _ := strconv.Atoi(isFindingBuddy)
	note := models.Note{
		NoteTitle:        noteTitle,
		NoteContent:      noteContent,
		ViewCount:        0,
		NoteTagList:      noteTagList,
		NoteType:         noteType,
		NoteURLs:         string(noteURLsJSON),
		NoteCreatorID:    uint(creatorID),
		NoteUpdateTime:   time.Now().Unix(),
		IsFindingBuddy:   isFindingBuddyInt,
		BuddyDescription: buddyDescription,
	}

	// 保存 Note 到数据库
	if err := global.Db.Create(&note).Error; err != nil {
		// 删除已上传的文件
		cleanupUploadedFiles(uploadedURLs)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "笔记保存失败",
		})
		return
	}

	// 更新用户的 NoteCount 字段
	if err := global.Db.Model(&models.User{}).
		Where("user_id = ?", creatorID).
		Update("note_count", gorm.Expr("note_count + ?", 1)).Error; err != nil {
		// 删除已上传的文件，并删除 Note 记录
		cleanupUploadedFiles(uploadedURLs)
		global.Db.Delete(&note)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "更新用户 NoteCount 失败",
		})
		return
	}

	// 处理 Tag 和 TagNoteRelation
	tags := strings.Split(noteTagList, ",")
	for _, tagName := range tags {
		var tag models.Tag
		if err := global.Db.Where("t_name = ?", tagName).First(&tag).Error; err != nil {
			tag = models.Tag{
				ID:         strconv.FormatInt(time.Now().UnixNano(), 10),
				TName:      tagName,
				Creator:    noteCreatorID,
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

		relation := models.TagNoteRelation{
			NID:        note.NoteID,
			TID:        tag.ID,
			CreatorID:  uint(creatorID),
			CreateDate: time.Now(),
		}
		global.Db.Create(&relation)
	}

	// 成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"nid":    note.NoteID,
		"urls":   uploadedURLs,
	})
}

// UpdateNoteWithPics 更新多图笔记接口（含相应更新oss上文件）
func UpdateNoteWithPics(ctx *gin.Context) {
	// 获取请求参数
	noteID := ctx.PostForm("note_id")
	noteTitle := ctx.PostForm("note_title")
	noteContent := ctx.PostForm("note_content")
	noteTagList := ctx.PostForm("note_tag_list")
	noteType := ctx.PostForm("note_type")
	isFindingBuddy := ctx.PostForm("is_finding_buddy")
	buddyDescription := ctx.PostForm("buddy_description")

	// 检查必要参数
	if noteID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要的 Note ID 参数",
		})
		return
	}

	if isFindingBuddy != "0" || isFindingBuddy != "1" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "isFindingBuddy参数只可以为0或1",
		})
		return
	}

	// 如果是找旅伴帖子，检查 buddy_description 是否为空
	if isFindingBuddy == "1" && buddyDescription == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "找旅伴帖子必须提供需求描述",
		})
		return
	}

	// 根据 NoteID 查找笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", noteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, UpdateNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}

	// 更新 Tag 和 TagNoteRelation
	tagList := strings.Split(noteTagList, ",") // 新的 tag 列表

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
	if noteTitle != "" {
		note.NoteTitle = noteTitle
	}
	if noteContent != "" {
		note.NoteContent = noteContent
	}
	if noteTagList != "" {
		note.NoteTagList = noteTagList
	}
	if noteType != "" {
		note.NoteType = noteType
	}
	if isFindingBuddy == "0" {
		isFindingBud, _ := strconv.Atoi(isFindingBuddy)
		note.IsFindingBuddy = isFindingBud
	}
	if isFindingBuddy == "1" && buddyDescription != "" {
		note.BuddyDescription = buddyDescription
	}

	// 存下旧的文件urls
	var oldURLs []string
	if err := json.Unmarshal([]byte(note.NoteURLs), &oldURLs); err != nil {
		ctx.JSON(http.StatusInternalServerError, UpdateNoteResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记更新失败:" + err.Error(),
		})
		return
	}

	// 上传新文件
	files := ctx.Request.MultipartForm.File["files"]
	var newUploadedURLs []string
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"status": "失败",
				"code":   400,
				"error":  "笔记更新文件不支持的文件类型",
			})
			return
		}

		url, err := oss.UploadFileToAliyunOss(file, "note_pics")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"status": "失败",
				"code":   500,
				"error":  "笔记更新文件上传至oss失败",
			})
			return
		}
		newUploadedURLs = append(newUploadedURLs, url)
	}

	// 更新 NoteURLs
	noteURLsJSON, err := json.Marshal(newUploadedURLs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "URL 序列化失败",
		})
		return
	}
	note.NoteURLs = string(noteURLsJSON)

	note.NoteUpdateTime = time.Now().Unix() // 更新时间戳

	// 保存到数据库
	if err := global.Db.Save(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "笔记更新失败",
		})
		return
	}

	cleanupUploadedFiles(oldURLs) // 安心删除旧文件

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"urls":   newUploadedURLs,
	})
}

// PublishNoteWithVideo 发布视频笔记（含上传笔记视频至oss）
func PublishNoteWithVideo(ctx *gin.Context) {
	// 接收笔记信息
	noteTitle := ctx.PostForm("note_title")
	noteContent := ctx.PostForm("note_content")
	noteTagList := ctx.PostForm("note_tag_list")
	noteType := ctx.PostForm("note_type")
	noteCreatorID := ctx.PostForm("note_creator_id")
	isFindingBuddy := ctx.PostForm("is_finding_buddy")
	buddyDescription := ctx.PostForm("buddy_description")

	if noteTitle == "" || noteContent == "" || noteCreatorID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要参数",
		})
		return
	}

	if isFindingBuddy != "0" || isFindingBuddy != "1" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "isFindingBuddy参数只可以为0或1",
		})
		return
	}

	// 如果是找旅伴帖子，检查 buddy_description 是否为空
	if isFindingBuddy == "1" && buddyDescription == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "找旅伴帖子必须提供需求描述",
		})
		return
	}

	creatorID, err := strconv.Atoi(noteCreatorID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "创建者ID格式错误",
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := global.Db.First(&user, "user_id = ?", creatorID).Error; err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "用户不存在",
		})
		return
	}

	// 处理文件
	videoFile, err := ctx.FormFile("video_file")
	if videoFile == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  "没有上传文件，或者请求格式不正确",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "视频文件上传失败",
		})
		return
	}
	// 校验视频格式与大小
	ext := strings.ToLower(filepath.Ext(videoFile.Filename))
	if ext != ".mp4" && ext != ".mov" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "不支持的视频格式，仅支持 mp4 和 mov",
		})
		return
	}
	if videoFile.Size > 20*1024*1024*1024 { // 20GB
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "视频文件大小超出限制，最大支持20GB",
		})
		return
	}

	var newVideoURLs []string

	// 上传新视频文件到 OSS
	newVideoURL, err := oss.UploadFileToAliyunOss(videoFile, "note_videos")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "视频上传失败",
		})
		return
	}
	newVideoURLs = append(newVideoURLs, newVideoURL)

	// 转换 URL 为 JSON
	noteURLsJSON, err := json.Marshal(newVideoURLs)
	if err != nil {
		// 删除已上传的文件
		cleanupUploadedFiles(newVideoURLs)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "视频URL序列化失败",
		})
		return
	}

	// 创建 Note 记录
	isFindingBuddyInt, _ := strconv.Atoi(isFindingBuddy)
	note := models.Note{
		NoteTitle:        noteTitle,
		NoteContent:      noteContent,
		ViewCount:        0,
		NoteTagList:      noteTagList,
		NoteType:         noteType,
		NoteURLs:         string(noteURLsJSON),
		NoteCreatorID:    uint(creatorID),
		NoteUpdateTime:   time.Now().Unix(),
		IsFindingBuddy:   isFindingBuddyInt,
		BuddyDescription: buddyDescription,
	}

	// 保存 Note 到数据库
	if err := global.Db.Create(&note).Error; err != nil {
		// 删除已上传的文件
		cleanupUploadedFiles(newVideoURLs)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "笔记保存失败",
		})
		return
	}

	// 更新用户的 NoteCount 字段
	if err := global.Db.Model(&models.User{}).
		Where("user_id = ?", creatorID).
		Update("note_count", gorm.Expr("note_count + ?", 1)).Error; err != nil {
		// 删除已上传的文件，并删除 Note 记录
		cleanupUploadedFiles(newVideoURLs)
		global.Db.Delete(&note)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "更新用户 NoteCount 失败",
		})
		return
	}

	// 处理 Tag 和 TagNoteRelation
	tags := strings.Split(noteTagList, ",")
	for _, tagName := range tags {
		var tag models.Tag
		if err := global.Db.Where("t_name = ?", tagName).First(&tag).Error; err != nil {
			tag = models.Tag{
				ID:         strconv.FormatInt(time.Now().UnixNano(), 10),
				TName:      tagName,
				Creator:    noteCreatorID,
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

		relation := models.TagNoteRelation{
			NID:        note.NoteID,
			TID:        tag.ID,
			CreatorID:  uint(creatorID),
			CreateDate: time.Now(),
		}
		global.Db.Create(&relation)
	}

	// 成功响应
	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"nid":    note.NoteID,
		"urls":   newVideoURLs,
	})
}

// UpdateNoteWithVideo 更新视频笔记接口（含相应更新oss上文件）
func UpdateNoteWithVideo(ctx *gin.Context) {
	// 获取请求参数
	noteID := ctx.PostForm("note_id")
	noteTitle := ctx.PostForm("note_title")
	noteContent := ctx.PostForm("note_content")
	noteTagList := ctx.PostForm("note_tag_list")
	noteType := ctx.PostForm("note_type")
	isFindingBuddy := ctx.PostForm("is_finding_buddy")
	buddyDescription := ctx.PostForm("buddy_description")

	// 检查必要参数
	if noteID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "缺少必要的 Note ID 参数",
		})
		return
	}

	if isFindingBuddy != "0" || isFindingBuddy != "1" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "isFindingBuddy参数只可以为0或1",
		})
		return
	}

	// 如果是找旅伴帖子，检查 buddy_description 是否为空
	if isFindingBuddy == "1" && buddyDescription == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "找旅伴帖子必须提供需求描述",
		})
		return
	}

	// 根据 NoteID 查找笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", noteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, UpdateNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}

	// 更新 Tag 和 TagNoteRelation
	tagList := strings.Split(noteTagList, ",") // 新的 tag 列表

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
	if noteTitle != "" {
		note.NoteTitle = noteTitle
	}
	if noteContent != "" {
		note.NoteContent = noteContent
	}
	if noteTagList != "" {
		note.NoteTagList = noteTagList
	}
	if noteType != "" {
		note.NoteType = noteType
	}
	if isFindingBuddy == "0" {
		isFindingBud, _ := strconv.Atoi(isFindingBuddy)
		note.IsFindingBuddy = isFindingBud
	}
	if isFindingBuddy == "1" && buddyDescription != "" {
		note.BuddyDescription = buddyDescription
	}

	// 存下旧的文件urls
	var oldURLs []string
	if err := json.Unmarshal([]byte(note.NoteURLs), &oldURLs); err != nil {
		ctx.JSON(http.StatusInternalServerError, UpdateNoteResponse{
			Status: "失败",
			Code:   500,
			Error:  "笔记更新失败:" + err.Error(),
		})
		return
	}

	// 处理文件
	videoFile, err := ctx.FormFile("video_file")
	if videoFile == nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Status: "失败",
			Code:   400,
			Error:  "没有上传文件，或者请求格式不正确",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "视频文件上传失败",
		})
		return
	}
	// 校验视频格式与大小
	ext := strings.ToLower(filepath.Ext(videoFile.Filename))
	if ext != ".mp4" && ext != ".mov" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "不支持的视频格式，仅支持 mp4 和 mov",
		})
		return
	}
	if videoFile.Size > 20*1024*1024*1024 { // 20GB
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "失败",
			"code":   400,
			"error":  "视频文件大小超出限制，最大支持20GB",
		})
		return
	}

	var newVideoURLs []string

	// 上传新视频文件到 OSS
	newVideoURL, err := oss.UploadFileToAliyunOss(videoFile, "note_videos")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "视频上传失败",
		})
		return
	}
	newVideoURLs = append(newVideoURLs, newVideoURL)

	// 更新 NoteURLs
	noteURLsJSON, err := json.Marshal(newVideoURLs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "URL 序列化失败",
		})
		return
	}
	note.NoteURLs = string(noteURLsJSON)

	note.NoteUpdateTime = time.Now().Unix() // 更新时间戳

	// 保存到数据库
	if err := global.Db.Save(&note).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "笔记更新失败",
		})
		return
	}

	cleanupUploadedFiles(oldURLs) // 安心删除旧文件

	ctx.JSON(http.StatusOK, gin.H{
		"status": "成功",
		"code":   200,
		"urls":   newVideoURLs,
	})
}

// DeleteNote 删除笔记接口（含相应删除oss上文件）
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

	// 查找笔记以获取创建者 ID
	var note models.Note
	if err := global.Db.Where("note_id = ?", req.NoteID).First(&note).Error; err == nil {
		// 更新创建者的 NoteCount 字段
		if err := global.Db.Model(&models.User{}).
			Where("user_id = ?", note.NoteCreatorID).
			Update("note_count", gorm.Expr("note_count - ?", 1)).Error; err != nil {
			fmt.Printf("[ERROR] Failed to update User's NoteCount for UserID %d: %s\n", note.NoteCreatorID, err)
			ctx.JSON(http.StatusInternalServerError, ErrorResponse{
				Status: "失败",
				Code:   500,
				Error:  "删除笔记失败:" + err.Error(),
			})
			return
		}
	} else {
		fmt.Printf("[ERROR] Failed to find Note with ID %d: %s\n", req.NoteID, err)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "删除笔记失败:" + err.Error(),
		})
		return
	}

	// 删除 Note
	if err := global.Db.Delete(&models.Note{}, req.NoteID).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Status: "失败",
			Code:   500,
			Error:  "删除笔记失败:" + err.Error(),
		})
		return
	}

	// 最后再删除oss笔记文件，调用 cleanupUploadedFiles 删除文件
	var uploadedURLs []string
	if err := json.Unmarshal([]byte(note.NoteURLs), &uploadedURLs); err != nil {
		log.Printf("解析 NoteURLs 失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status": "失败",
			"code":   500,
			"error":  "解析 NoteURLs 失败",
		})
		return
	}
	cleanupUploadedFiles(uploadedURLs)

	// 成功响应
	ctx.JSON(http.StatusOK, DeleteNoteResponse{
		Status: "笔记删除成功",
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.FormatInt(notes[len(notes)-1].NoteUpdateTime, 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}

// GetLikedNotes 获取用户点赞的帖子
func GetLikedNotes(ctx *gin.Context) {
	userID := ctx.Query("user_id")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor")

	if userID == "" || num == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "参数缺失",
		})
		return
	}

	limit := 30
	if n, err := strconv.Atoi(num); err == nil && n > 0 && n < 30 {
		limit = n
	}

	query := global.Db.Where("uid = ?", userID)
	if cursor != "" {
		if timestamp, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			query = query.Where("create_date < ?", time.Unix(timestamp, 0))
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标参数",
			})
			return
		}
	}

	var likes []models.Like
	if err := query.Order("create_date DESC").Limit(limit).Find(&likes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询点赞记录失败",
		})
		return
	}

	var noteIDs []uint
	for _, like := range likes {
		noteIDs = append(noteIDs, like.Nid)
	}

	var notes []models.Note
	if err := global.Db.Where("note_id IN ?", noteIDs).Find(&notes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询笔记失败",
		})
		return
	}

	var responseNotes []gin.H
	var nextCursor string
	for _, note := range notes {
		responseNotes = append(responseNotes, gin.H{
			"note_id":          note.NoteID,
			"note_title":       note.NoteTitle,
			"note_content":     note.NoteContent,
			"like_counts":      note.LikeCounts,
			"collect_counts":   note.CollectCounts,
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	if len(likes) > 0 {
		nextCursor = strconv.FormatInt(likes[len(likes)-1].CreateDate.Unix(), 10)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor,
		},
	})
}

// GetCollectedNotes 获取用户收藏的帖子
func GetCollectedNotes(ctx *gin.Context) {
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

	// 构造查询条件
	query := global.Db.Where("uid = ?", userID)
	if cursor != "" {
		// 游标为时间戳（Unix 时间）
		if timestamp, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			query = query.Where("create_date < ?", time.Unix(timestamp, 0))
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标参数",
			})
			return
		}
	}

	// 查询收藏记录
	var collects []models.Collect
	if err := query.Order("create_date DESC").Limit(limit).Find(&collects).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询收藏记录失败",
		})
		return
	}

	// 获取收藏的笔记详情
	var noteIDs []uint
	for _, collect := range collects {
		noteIDs = append(noteIDs, collect.Nid)
	}

	var notes []models.Note
	if err := global.Db.Where("note_id IN ?", noteIDs).Find(&notes).Error; err != nil {
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(collects) > 0 {
		nextCursor = strconv.FormatInt(collects[len(collects)-1].CreateDate.Unix(), 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    0,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor,
		},
	})
}

// GetNoteByID 根据笔记 ID 获取笔记信息
func GetNoteByID(ctx *gin.Context) {
	// 从路径参数获取 note_id
	noteID := ctx.Query("note_id")
	if noteID == "" {
		ctx.JSON(http.StatusBadRequest, GetNoteResponse{
			Status: "失败",
			Code:   400,
			Error:  "笔记id不得为空",
		})
		return
	}

	// 查询数据库中是否存在该笔记
	var note models.Note
	if err := global.Db.First(&note, "note_id = ?", noteID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, GetNoteResponse{
			Status: "失败",
			Code:   404,
			Error:  "笔记不存在",
		})
		return
	}
	// 假设 noteURLs 是存储 JSON 字符串的字段
	var noteURLs []string
	if err := json.Unmarshal([]byte(note.NoteURLs), &noteURLs); err != nil {
		ctx.JSON(http.StatusInternalServerError, GetNoteResponse{
			Status: "失败",
			Code:   500,
			Error:  "解析 noteURLs 失败",
		})
		return
	}
	// 返回笔记数据
	ctx.JSON(http.StatusOK, GetNoteResponse{
		Status:         "成功",
		Code:           200,
		NoteID:         note.NoteID,
		NoteTitle:      note.NoteTitle,
		NoteContent:    note.NoteContent,
		ViewCount:      uint(int(note.ViewCount)),
		NoteTagList:    note.NoteTagList,
		NoteType:       note.NoteType,
		NoteCreatorID:  note.NoteCreatorID,
		NoteUpdateTime: note.NoteUpdateTime,
		LikeCounts:     uint(note.LikeCounts),
		CollectCounts:  uint(int(note.CollectCounts)),
		NoteURLS:       noteURLs,
	})
}

// GetNotesByCreatorID 根据创建者 ID 获取笔记
func GetNotesByCreatorID(ctx *gin.Context) {
	// 获取请求参数
	creatorId := ctx.Query("creator_id")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（时间戳）

	// 参数校验
	if creatorId == "" || num == "" {
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

	// 根据游标查询
	query := global.Db
	if cursor != "" {
		// 使用游标（时间戳）来进行分页，获取小于游标的记录（倒序）
		cursorTime, err := strconv.ParseInt(cursor, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标",
			})
			return
		}
		// 游标小于等于当前时间的记录进行查询
		query = query.Where("note_update_time < ?", cursorTime)
	} else {
		// 如果没有提供游标，获取所有数据
		query = query
	}

	// 查询帖子数据，按照更新的时间倒序排序，返回最多 `limit` 条
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.FormatInt(notes[len(notes)-1].NoteUpdateTime, 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}

// GetNotesByUpdateTime 根据更新时间新旧获取笔记
func GetNotesByUpdateTime(ctx *gin.Context) {
	// 获取请求参数
	noteType := ctx.Query("note_type")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（时间戳）

	// 参数校验
	if num == "" {
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

	// 构造查询条件
	query := global.Db
	if noteType != "" {
		query = query.Where("note_type = ?", noteType)
	}
	if cursor != "" {
		if timestamp, err := strconv.ParseInt(cursor, 10, 64); err == nil {
			query = query.Where("note_update_time < ?", timestamp) // 返回最近更新的帖子
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.FormatInt(notes[len(notes)-1].NoteUpdateTime, 10)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}

// GetNotesByLikes 根据笔记获赞数多少获取笔记
func GetNotesByLikes(ctx *gin.Context) {
	// 获取请求参数
	noteType := ctx.Query("note_type")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（点赞数）

	// 参数校验
	if num == "" {
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

	// 构造查询条件
	query := global.Db
	if noteType != "" {
		query = query.Where("note_type = ?", noteType)
	}
	if cursor != "" {
		if likes, err := strconv.Atoi(cursor); err == nil && likes >= 0 {
			query = query.Where("note_like < ?", likes) // 返回点赞数较低的记录
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
	if err := query.Order("note_like DESC").Limit(limit).Find(&notes).Error; err != nil {
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.Itoa(notes[len(notes)-1].LikeCounts)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}

func GetNotesByCollects(ctx *gin.Context) {
	// 获取请求参数
	noteType := ctx.Query("note_type")
	num := ctx.Query("num")
	cursor := ctx.Query("cursor") // 游标，用于分页（点赞数）

	// 参数校验
	if num == "" {
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

	// 构造查询条件
	query := global.Db
	if noteType != "" {
		query = query.Where("note_type = ?", noteType)
	}
	if cursor != "" {
		if likes, err := strconv.Atoi(cursor); err == nil && likes >= 0 {
			query = query.Where("collect_counts < ?", likes) // 返回点赞数较低的记录
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
	if err := query.Order("collect_counts DESC").Limit(limit).Find(&notes).Error; err != nil {
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
			"comment_counts":   note.CommentCounts,
			"note_creator_id":  note.NoteCreatorID,
			"note_update_time": note.NoteUpdateTime,
			"note_type":        note.NoteType,
			"note_tag_list":    note.NoteTagList,
			"view_count":       note.ViewCount,
			"note_urls":        note.NoteURLs,
		})
	}

	// 设置下一个游标
	if len(notes) > 0 {
		nextCursor = strconv.Itoa(notes[len(notes)-1].LikeCounts)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":    200,
		"success": true,
		"msg":     "成功",
		"data": gin.H{
			"notes":      responseNotes,
			"nextCursor": nextCursor, // 下次分页使用的游标
		},
	})
}

// GetHotRecommendationsResponse 响应结构体
type GetHotRecommendationsResponse struct {
	Code    int                 `json:"code"`
	Success bool                `json:"success"`
	Msg     string              `json:"msg"`
	Data    RecommendationsData `json:"data"`
}

// RecommendationsData 数据部分
type RecommendationsData struct {
	Notes      []NoteResponse `json:"notes"`
	NextCursor string         `json:"nextCursor"`
}

// NoteResponse 笔记返回结构体
type NoteResponse struct {
	NoteID         uint    `json:"note_id"`
	NoteTitle      string  `json:"note_title"`
	NoteContent    string  `json:"note_content"`
	LikeCounts     uint    `json:"like_counts"`
	CollectCounts  uint    `json:"collect_counts"`
	CommentCounts  uint    `json:"comment_counts"`
	NoteCreatorID  uint    `json:"note_creator_id"`
	NoteUpdateTime uint    `json:"note_update_time"`
	ViewCount      uint    `json:"view_count"`
	NoteTagList    string  `json:"note_tag_list"`
	NoteURLs       string  `json:"note_urls"`
	Score          float64 `json:"score"` // 热度分数
}

// 获取热度推荐
func GetHotRecommendations(ctx *gin.Context) {
	// 获取请求参数
	numStr := ctx.Query("num")
	cursorStr := ctx.Query("cursor") // 游标，用于分页（基于分数）

	// 参数校验
	if numStr == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "参数缺失",
		})
		return
	}

	// 默认最大条数
	limit := 30
	if n, err := strconv.Atoi(numStr); err == nil && n > 0 && n <= 30 {
		limit = n
	} else if err != nil || n <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"success": false,
			"msg":     "无效的num参数",
		})
		return
	}

	// 构造查询条件
	query := global.Db.Model(&models.Note{})

	// 如果有游标，添加过滤条件，只基于score进行分页
	if cursorStr != "" {
		cursorScore, err := strconv.ParseFloat(cursorStr, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"success": false,
				"msg":     "无效的游标参数",
			})
			return
		}
		// 基于分数进行分页，score < 游标分数
		query = query.Where("score < ?", cursorScore)
	}

	// 查询笔记数据，按分数降序排序
	var notes []models.Note
	if err := query.Order("score DESC").Limit(limit).Find(&notes).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"success": false,
			"msg":     "查询笔记失败",
		})
		return
	}

	// 构造返回结果
	var responseNotes []NoteResponse
	for _, note := range notes {
		responseNotes = append(responseNotes, NoteResponse{
			NoteID:         note.NoteID,
			NoteTitle:      note.NoteTitle,
			NoteContent:    note.NoteContent,
			LikeCounts:     uint(note.LikeCounts),
			CollectCounts:  note.CollectCounts,
			CommentCounts:  note.CommentCounts,
			NoteCreatorID:  note.NoteCreatorID,
			NoteUpdateTime: uint(note.NoteUpdateTime),
			ViewCount:      note.ViewCount,
			NoteTagList:    note.NoteTagList,
			NoteURLs:       note.NoteURLs,
			Score:          note.Score,
		})
	}

	// 设置下一个游标
	var nextCursor string
	if len(notes) == limit {
		lastNote := notes[len(notes)-1]
		nextCursor = strconv.FormatFloat(lastNote.Score, 'f', 6, 64) // 将最后一个笔记的score作为游标
	}

	// 构造响应
	response := GetHotRecommendationsResponse{
		Code:    200,
		Success: true,
		Msg:     "成功",
		Data: RecommendationsData{
			Notes:      responseNotes,
			NextCursor: nextCursor, // 下次分页使用的游标
		},
	}

	// 返回结果
	ctx.JSON(http.StatusOK, response)
}
