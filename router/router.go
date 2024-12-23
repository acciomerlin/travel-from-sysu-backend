package router

//路由管理文件，处理方法移步controllers

import (
	"travel-from-sysu-backend/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	auth := r.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/changePwd", controllers.ChangePwd)
		auth.POST("/changeUserInfo", controllers.ChangeUserInfo)
		auth.GET("/getUserInfoByID", controllers.GetUserInfoByID)
		auth.POST("/uploadAvatar", controllers.UploadAvatar)
		auth.GET("/getAvatar", controllers.GetAvatar)
	}
	note := r.Group("/api/note")
	{
		note.POST("/uploadNotePic", controllers.UploadNotePic)
		note.POST("/publishNoteWithPics", controllers.PublishNoteWithPics)
		note.POST("/updateNoteWithPics", controllers.UpdateNoteWithPics)
		note.POST("/uploadNoteVideo", controllers.UploadNoteVideo)
		note.POST("/publishNoteWithVideo", controllers.PublishNoteWithVideo)
		note.POST("/updateNoteWithVideo", controllers.UpdateNoteWithVideo)
		note.GET("/deleteUploadedFile", controllers.DeleteUploadedFile)
		note.POST("/deleteNote", controllers.DeleteNote)
		note.POST("/like", controllers.Like)
		note.POST("/dislike", controllers.Dislike)
		note.POST("/collect", controllers.Collect)
		note.POST("/uncollect", controllers.Uncollect)
		note.GET("/getNoteById", controllers.GetNoteByID)
		note.GET("/getNotesByCreatorId", controllers.GetNotesByCreatorID)
		note.GET("/getUserFoNotes", controllers.GetFoNotes)
		note.GET("/getNotesByUpdateTime", controllers.GetNotesByUpdateTime)
		note.GET("/getUserLikeNotes", controllers.GetLikedNotes)
		note.GET("/getUserCollectNotes", controllers.GetCollectedNotes)
		note.GET("/getNotesByLikes", controllers.GetNotesByLikes)
		note.GET("/getNotesByCollects", controllers.GetNotesByCollects)
		note.GET("/getHotRecommendations", controllers.GetHotRecommendations)
		note.GET("/getNotesByTag", controllers.GetNotesByTag)
		note.GET("/getNotesByKeywords", controllers.GetNoteByKeywords)
		note.GET("getIfUserFollow", controllers.GetIfUserFollow)
		note.GET("getIfUserLikeOrCollect", controllers.GetIfUserLikeOrCollect)
	}
	user := r.Group("/api/user")
	{
		user.POST("/follow", controllers.Follow)
		user.POST("/unfollow", controllers.Unfollow)
		user.GET("/getUserFoCounts", controllers.GetUserFoCounts)
		user.GET("/getFollowees", controllers.GetFolloweesWithPagination)
		user.GET("/getFollowers", controllers.GetFollowersWithPagination)
		user.GET("/getUserNoteCounts", controllers.GetNoteCountsByID)
	}
	comment := r.Group("/api/comment")
	{
		comment.POST("/deleteComment", controllers.DeleteComment)
		comment.POST("/publishComment", controllers.PublishComment)
		comment.GET("/getCommentById", controllers.GetCommentById)
		comment.GET("/getFirstLevelCommentsByNoteId", controllers.GetFirstLevelCommentsByNoteId)
		comment.GET("/getSecondLevelCommentsByParentId", controllers.GetSecondLevelCommentsByParentId)

	}
	notification := r.Group("/api/notification")
	{
		// 未读消息相关路由
		notification.GET("/unread_noti_count", controllers.GetUnreadNotificationCount)                   // 获取未读消息计数
		notification.GET("/unread_comments", controllers.GetUnreadCommentNotifications)                  // 获取未读评论消息
		notification.GET("/unread_likes-and-collects", controllers.GetUnreadLikeAndCollectNotifications) // 获取未读点赞+收藏消息
		notification.GET("/unread_follows", controllers.GetNewFollowNotifications)                       // 获取新增关注消息

		// 历史已读消息相关路由
		notification.GET("/read_comments", controllers.GetReadCommentNotifications)                  // 获取已读评论消息
		notification.GET("/read_likes-and-collects", controllers.GetReadLikeAndCollectNotifications) // 获取已读点赞+收藏消息
		notification.GET("/read_follows", controllers.GetReadFollowNotifications)                    // 获取已读关注消息
	}
	return r
}
