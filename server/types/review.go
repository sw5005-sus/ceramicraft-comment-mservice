package types

import "time"

type CreateReviewRequest struct {
	ProductID   int
	Content     string
	ParentID    string
	Stars       int
	PicInfo     []string `json:"pic_info"`
	IsAnonymous bool     `json:"is_anonymous"`
}

type LikeRequest struct {
	ReviewID string `json:"review_id"`
}

type ReviewInfo struct {
	ID               string    `json:"id"`
	Content          string    `json:"content"`
	UserID           int       `json:"user_id"`
	ProductID        int       `json:"product_id"`
	ParentID         string    `json:"parent_id"`
	Stars            int       `json:"stars"`
	IsAnonymous      bool      `json:"is_anonymous"`
	PicInfo          []string  `json:"pic_info"`
	CreatedAt        time.Time `json:"created_at"`
	Likes            int       `json:"likes"`
	CurrentUserLiked bool      `json:"current_user_liked"`
	IsPinned         bool      `json:"is_pinned"`
}

type PinReviewRequest struct {
	IsPinned bool   `json:"is_pinned"`
}

type ListReviewResponse struct {
	ReviewList   []ReviewInfo `json:"review_list"`
	PinnedReview *ReviewInfo  `json:"pinned_review"`
}

type ListReviewRequest struct {
	ProductID int `json:"product_id"`
	Stars     int `json:"stars"` // 0 means any stars
}

type UpdateReviewStatusRequest struct {
	ReviewID  string `json:"review_id"`           // required
	Status    string `json:"status"`              // required: pending / processing / approved / hidden / rejected
	IsMismatch *bool  `json:"is_mismatch,omitempty"` // optional
	IsHarmful *bool  `json:"is_harmful,omitempty"`   // optional
	AutoFlag  *string `json:"auto_flag,omitempty"`    // optional
}
