package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/http/data"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/service"
	"github.com/sw5005-sus/ceramicraft-comment-mservice/server/types"
)

// CreateReview Create.
//
// @Summary Review Create
// @Description Create an review record.
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.CreateReviewRequest true "CreateReviewRequest"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200	{object} data.BaseResponse{data=types.CreateReviewRequest}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/reviews [post]
func CreateReview(c *gin.Context) {
	var req types.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	userID := c.Value("userID").(int)
	err := service.GetReviewServiceInstance().CreateReview(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "create review success"))
}

// Like review
// @Summary Like a review
// @Description Like a review by id
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.LikeRequest true "LikeRequest"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=string}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/reviews/{review_id}/like [post]
func Like(c *gin.Context) {
	reviewID := c.Param("review_id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "empty review_id"})
		return
	}
	var req types.LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	req.ReviewID = reviewID
	userID := c.Value("userID").(int)
	err := service.GetReviewServiceInstance().Like(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "like success"))
}

// Get reviews by current user
// @Summary Get reviews by user
// @Description Get list of reviews created by current authenticated user
// @Tags Review
// @Accept json
// @Produce json
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=[]types.ReviewInfo}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/reviews/user [get]
func GetListByUserID(c *gin.Context) {
	userID := c.Value("userID").(int)
	list, err := service.GetReviewServiceInstance().GetListByUserID(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, list))
}

// Get reviews by product id
// @Summary Get reviews by product
// @Description Get list of reviews for a product
// @Tags Review
// @Accept json
// @Produce json
// @Param product_id path int true "Product ID"
// @Param client path string true "Client identifier" Enums(customer, merchant)
// @Success 200 {object} data.BaseResponse{data=types.ListReviewResponse}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/customer/reviews/product/{product_id} [get]
func GetListByProductID(c *gin.Context) {
	pidStr := c.Param("product_id")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "invalid product_id"})
		return
	}
	userID := 0
	if v := c.Value("userID"); v != nil {
		userID = v.(int)
	}
	list, err := service.GetReviewServiceInstance().GetListByProductID(c, pid, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, list))
}

// ListReviewsByFilter
// @Summary List reviews by product and stars
// @Description Filter reviews by product_id and stars (0 means any), ordered by created_at desc
// @Tags Review
// @Accept json
// @Produce json
// @Param filter body types.ListReviewRequest true "ListReviewRequest"
// @Success 200 {object} data.BaseResponse{data=[]types.ReviewInfo}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/list [post]
func ListReviewsByFilter(c *gin.Context) {
	var req types.ListReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	userID := 0
	if v := c.Value("userID"); v != nil {
		userID = v.(int)
	}
	resp, err := service.GetReviewServiceInstance().GetListByQuery(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, resp))
}

// Pin review
// @Summary Pin a review
// @Description Pin a review by id
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.PinReviewRequest true "PinReviewRequest"
// @Success 200 {object} data.BaseResponse{data=string}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/{review_id} [patch]
func PinReview(c *gin.Context) {
	reviewID := c.Param("review_id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "empty review_id"})
		return
	}
	var req types.PinReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	if req.IsPinned {
		err := service.GetReviewServiceInstance().PinReview(c, reviewID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
			return
		}
		c.JSON(http.StatusOK, RespSuccess(c, "pin success"))
		return
	}

	c.JSON(http.StatusOK, RespSuccess(c, nil))
}

// Delete review
// @Summary Delete a review
// @Description Delete a review by id
// @Tags Review
// @Accept json
// @Produce json
// @Param review_id path string true "Review ID"
// @Success 200 {object} data.BaseResponse{data=string}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/{review_id} [delete]
func DeleteReview(c *gin.Context) {
	reviewID := c.Param("review_id")
	if reviewID == "" {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "empty review_id"})
		return
	}
	err := service.GetReviewServiceInstance().DeleteReview(c, reviewID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "delete success"))
}

// UpdateReviewStatus Update review moderation status and fields.
//
// @Summary Update Review
// @Description Update review fields: status (required), is_mismatch/is_harmful/auto_flag (optional). For agent service use.
// @Tags Review
// @Accept json
// @Produce json
// @Param request body types.UpdateReviewStatusRequest true "UpdateReviewStatusRequest"
// @Success 200 {object} data.BaseResponse{data=string}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/status [post]
func UpdateReviewStatus(c *gin.Context) {
	var req types.UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	err := service.GetReviewServiceInstance().UpdateReview(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "update review success"))
}

// GetListByStatus Get list by status.
//
// @Summary Get reviews by status
// @Description Get all reviews with specified status.
// @Tags Review
// @Accept json
// @Produce json
// @Param status path string true "Review status" Enums(pending, processing, approved, hidden, rejected)
// @Success 200 {object} data.BaseResponse{data=[]types.ReviewInfo}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/status/{status} [get]
func GetListByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "status parameter is required"})
		return
	}

	list, err := service.GetReviewServiceInstance().GetListByStatus(c, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}

	c.JSON(http.StatusOK, RespSuccess(c, list))
}

// ReplyReview Reply.
//
// @Summary Review Reply
// @Description Reply an review record.
// @Tags Review
// @Accept json
// @Produce json
// @Param user body types.CreateReviewRequest true "CreateReviewRequest"
// @Success 200	{object} data.BaseResponse{data=types.CreateReviewRequest}
// @Failure 400 {object} data.BaseResponse{data=string}
// @Failure 500 {object} data.BaseResponse{data=string}
// @Router /comment-ms/v1/merchant/reviews/{review_id}/replies [post]
func ReplyReview(c *gin.Context) {
	parentID := c.Param("review_id")
	if parentID == "" {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: "empty review_id"})
		return
	}
	var req types.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	req.ParentID = parentID
	userID := c.Value("userID").(int)
	err := service.GetReviewServiceInstance().CreateReview(c, req, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, data.BaseResponse{ErrMsg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, RespSuccess(c, "reply review success"))
}
