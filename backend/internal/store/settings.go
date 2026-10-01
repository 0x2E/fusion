package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/0x2E/fusion/internal/model"
)

// GetSettings returns the singleton settings row. A missing row is not an
// error: it means no preference was ever saved, so all fields stay nil.
func (s *Store) GetSettings() (*model.Settings, error) {
	settings := &model.Settings{}
	err := s.db.QueryRow(`
		SELECT locale, article_page_size, theme, auto_mark_read, updated_at
		FROM settings
		WHERE id = 1
	`).Scan(&settings.Locale, &settings.ArticlePageSize, &settings.Theme, &settings.AutoMarkRead, &settings.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return settings, nil
		}
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return settings, nil
}

// UpdateSettingsParams supports partial updates. Only non-nil fields will be
// updated; nil means "leave unchanged".
type UpdateSettingsParams struct {
	Locale          *string
	ArticlePageSize *int64
	Theme           *string
	AutoMarkRead    *string
}

// UpdateSettings upserts the singleton settings row. Fields set to nil keep
// their stored value via COALESCE (excluded.* is the proposed insert row, so a
// NULL there means "not provided in this request"). A nil pointer argument is
// passed through to SQL as NULL by database/sql.
func (s *Store) UpdateSettings(params UpdateSettingsParams) error {
	if params.Locale == nil && params.ArticlePageSize == nil && params.Theme == nil &&
		params.AutoMarkRead == nil {
		return nil
	}

	_, err := s.db.Exec(`
		INSERT INTO settings (id, locale, article_page_size, theme, auto_mark_read)
		VALUES (1, :locale, :article_page_size, :theme, :auto_mark_read)
		ON CONFLICT (id) DO UPDATE SET
			locale = COALESCE(excluded.locale, settings.locale),
			article_page_size = COALESCE(excluded.article_page_size, settings.article_page_size),
			theme = COALESCE(excluded.theme, settings.theme),
			auto_mark_read = COALESCE(excluded.auto_mark_read, settings.auto_mark_read),
			updated_at = unixepoch()
	`,
		sql.Named("locale", params.Locale),
		sql.Named("article_page_size", params.ArticlePageSize),
		sql.Named("theme", params.Theme),
		sql.Named("auto_mark_read", params.AutoMarkRead),
	)
	if err != nil {
		return fmt.Errorf("update settings: %w", err)
	}
	return nil
}
