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
		auth.GET("/getNameByID", controllers.GetNameByID)
	}
	note := r.Group("/api/note")
	{
		note.POST("/publishNote", controllers.PublishNote)
		note.POST("/deleteNote", controllers.DeleteNote)
		note.POST("/updateNote", controllers.UpdateNote)
		note.GET("/getNoteById", controllers.GetNoteByID)
		note.GET("/getNotesByCreatorId", controllers.GetNotesByCreatorID)
		note.GET("/getUserFoNotes", controllers.GetFoNotes)
	}
	user := r.Group("/api/user")
	{
		user.POST("/follow", controllers.Follow)
		user.POST("/unfollow", controllers.Unfollow)
		user.GET("/getUserFoCounts", controllers.GetUserFoCounts)
		user.GET("/getFollowees", controllers.GetFolloweesWithPagination)
		user.GET("/getFollowers", controllers.GetFollowersWithPagination)
	}
	comment := r.Group("/api/comment")
	{
		comment.POST("/deleteComment", controllers.DeleteComment)
		comment.POST("/publishComment", controllers.PublishComment)
		comment.GET("/getCommentById", controllers.GetCommentById)
		comment.GET("/getFirstLevelCommentsByNoteId", controllers.GetFirstLevelCommentsByNoteId)
		comment.GET("/getSecondLevelCommentsByParentId", controllers.GetSecondLevelCommentsByParentId)
	}
	return r
}
