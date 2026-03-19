package router

import (
	"github.com/gin-gonic/gin"

	_ "github.com/sw5005-sus/ceramicraft-comment-mservice/server/docs"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/http/api"
	"github.com/sw5005-sus/ceramicraft-user-mservice/common/middleware"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

const (
	serviceURIPrefix = "/comment-ms/v1"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	basicGroup := r.Group(serviceURIPrefix)
	{
		basicGroup.GET("/swagger/*any", gs.WrapHandler(
			swaggerFiles.Handler,
			gs.URL("/comment-ms/v1/swagger/doc.json"),
		))
		basicGroup.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})
	}

	merchantGroup := basicGroup.Group("/merchant")
	{
		merchantGroup.Use(middleware.AuthMiddleware())
		merchantGroup.PATCH("/reviews/:review_id", middleware.RequireRoles("merchant_admin"), api.PinReview)
		merchantGroup.DELETE("/review/:review_id", middleware.RequireRoles("merchant_admin"), api.DeleteReview)
		merchantGroup.POST("/reviews/list", api.ListReviewsByFilter)
		merchantGroup.POST("/reviews/:review_id/replies", middleware.RequireRoles("merchant_admin"), api.ReplyReview)
	}

	customerGroup := basicGroup.Group("/customer")
	{
		customerGroup.Use(middleware.AuthMiddleware())
		customerGroup.POST("/reviews", api.CreateReview)
		customerGroup.POST("/reviews/:review_id/like", api.Like)
		customerGroup.GET("/reviews/user", api.GetListByUserID)
		customerGroup.GET("/reviews/product/:product_id", api.GetListByProductID)
	}
	return r
}
