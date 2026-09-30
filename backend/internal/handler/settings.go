package handler

import (
	"fmt"
	"slices"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

// These allow-lists mirror the SQL CHECK constraints and the frontend
// constants (supportedLocales / articlePageSizeOptions in the preferences
// store, and the theme literals in the settings dialog).
var (
	validLocales          = []string{"en", "zh", "de", "fr", "es", "ru", "pt", "sv"}
	validArticlePageSizes = []int64{10, 20, 30, 50, 100}
	validThemes           = []string{"light", "dark", "system"}
)

// A missing field and an explicit null are indistinguishable after JSON
// decoding (both leave the pointer nil) and both mean "leave unchanged".
type updateSettingsRequest struct {
	Locale          *string `json:"locale"`
	ArticlePageSize *int64  `json:"article_page_size"`
	Theme           *string `json:"theme"`
}

func (h *Handler) getSettings(c *gin.Context) {
	settings, err := h.store.GetSettings()
	if err != nil {
		internalError(c, err, "get settings")
		return
	}

	dataResponse(c, settings)
}

func (h *Handler) updateSettings(c *gin.Context) {
	var req updateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequestError(c, "invalid request")
		return
	}

	if err := validateSettingsUpdate(&req); err != nil {
		badRequestError(c, err.Error())
		return
	}

	if err := h.store.UpdateSettings(store.UpdateSettingsParams{
		Locale:          req.Locale,
		ArticlePageSize: req.ArticlePageSize,
		Theme:           req.Theme,
	}); err != nil {
		internalError(c, err, "update settings")
		return
	}

	settings, err := h.store.GetSettings()
	if err != nil {
		internalError(c, err, "get settings after update")
		return
	}

	dataResponse(c, settings)
}

func validateSettingsUpdate(req *updateSettingsRequest) error {
	if req.Locale != nil && !slices.Contains(validLocales, *req.Locale) {
		return fmt.Errorf("invalid locale")
	}
	if req.ArticlePageSize != nil && !slices.Contains(validArticlePageSizes, *req.ArticlePageSize) {
		return fmt.Errorf("invalid article_page_size")
	}
	if req.Theme != nil && !slices.Contains(validThemes, *req.Theme) {
		return fmt.Errorf("invalid theme")
	}
	return nil
}
