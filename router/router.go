package router

import (
	"ListCharacterGI/middleware"
	giService "ListCharacterGI/service"
	"github.com/gin-gonic/gin"
)

type RouteBuilder struct {
	dataService *giService.Handler
}

func NewRouteBuilder(dataService *giService.Handler) *RouteBuilder {
	return &RouteBuilder{dataService: dataService}
}

func (rb *RouteBuilder) RouteInit() *gin.Engine {

	r := gin.Default()
	r.Use(middleware.RecoveryMiddleware(), middleware.Logger())
	//r.Use(gin.Recovery(), middleware.Logger(), middleware.BasicAuth())

	auth := r.Group("/", middleware.AdminMiddleware())
	{
		auth.GET("/admin/viewUsers", rb.dataService.ViewAllUsers)
		auth.DELETE("admin/users/:user_id", rb.dataService.DeleteUser)
		auth.GET("/access", rb.dataService.Access)
		auth.POST("/create-characters", rb.dataService.HandlerCreate)
		auth.PUT("/listGenshin/:id", rb.dataService.HandlerUpdate)
		auth.DELETE("/listGenshin/:id", rb.dataService.HandlerDelete)
		auth.POST("/uploadS3/:id", rb.dataService.UploadFileS3AtchHandler)
		auth.POST("/uploadLocal/:id", rb.dataService.UploadLocalAtchHandler)
		auth.GET("/SearchCharacter", rb.dataService.SearchHandler)
		auth.GET("/download-attachment/:id", rb.dataService.DownloadAttachmentHandler)
		// message

	}
	r.GET("/get-message", middleware.UserMiddleware(), rb.dataService.HandlerGetMessages)
	r.POST("/send-message", middleware.UserMiddleware(), rb.dataService.HandlerSendMessage)

	r.GET("/", middleware.UserMiddleware(), rb.dataService.HandlerGetAll)
	r.GET("/listGenshin/:id", middleware.UserMiddleware(), rb.dataService.HandlerGetByID)

	// test
	r.POST("/uploadBuckets", rb.dataService.UploadFileS3BucketsHandler)

	r.POST("/register", rb.dataService.Register)
	r.POST("/login", rb.dataService.Login)
	return r
}
