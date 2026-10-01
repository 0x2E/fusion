package handler

import (
	"fmt"
	"slices"

	"github.com/0x2E/fusion/internal/store"
	"github.com/gin-gonic/gin"
)

// Locale and theme are closed sets by contract (shipped catalogs and
// next-themes' preference names). The page size is a free value within the
// range the items API can serve (maxListLimit); the five-option list is only
// a frontend UI choice, so a future custom-value picker needs no backend
// change. These allow-lists mirror the frontend constants in the preferences
// store and the settings dialog.
var (
	validLocales = []string{"en", "zh", "de", "fr", "es", "ru", "pt", "sv"}
	validThemes  = []string{"light", "dark", "system"}
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
	if req.ArticlePageSize != nil &&
		(*req.ArticlePageSize < 1 || *req.ArticlePageSize > maxListLimit) {
		return fmt.Errorf("invalid article_page_size")
	}
	if req.Theme != nil && !slices.Contains(validThemes, *req.Theme) {
		return fmt.Errorf("invalid theme")
	}
	return nil
}
