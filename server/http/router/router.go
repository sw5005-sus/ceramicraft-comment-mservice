package router

import (
	"github.com/gin-gonic/gin"

	auditclient "github.com/sw5005-sus/ceramicraft-audit-client"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/config"
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
	audit_middleware := auditclient.AuditMiddleware(
		"comment-ms",
		config.Config.AuditGrpcConfig.Host,
		config.Config.AuditGrpcConfig.Port)
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
		// Internal API for agent service - no auth required
		basicGroup.POST("/merchant/reviews/status", audit_middleware, api.UpdateReviewStatus)
		basicGroup.GET("/merchant/reviews/by-status", audit_middleware, api.GetListByStatus)
	}

	merchantGroup := basicGroup.Group("/merchant")
	{
		merchantGroup.Use(middleware.AuthMiddleware())
		merchantGroup.PATCH("/reviews/:review_id", middleware.RequireRoles("merchant_admin"), audit_middleware, api.PinReview)
		merchantGroup.DELETE("/reviews/:review_id", middleware.RequireRoles("merchant_admin"), audit_middleware, api.DeleteReview)
		merchantGroup.POST("/reviews/list", api.ListReviewsByFilter)
		merchantGroup.POST("/reviews/:review_id/replies", middleware.RequireRoles("merchant_admin"), audit_middleware, api.ReplyReview)
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
