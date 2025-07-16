package sqlrepo

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

// TranslationRepo implements TranslationRepository
type TranslationRepo struct {
	db *sql.DB
}

// NewTranslationRepo creates a new translation repository
func NewTranslationRepo(db *sql.DB) *TranslationRepo {
	return &TranslationRepo{db: db}
}

// Create creates a new translation
func (r *TranslationRepo) Create(ctx context.Context, translation *model.Translation) error {
	db := sql.GetDBOrTx(ctx, r.db)
	if err := db.Create(translation).Error; err != nil {
		return fmt.Errorf("create translation: %w", err)
	}
	return nil
}

// GetByID retrieves a translation by ID
func (r *TranslationRepo) GetByID(ctx context.Context, id string) (*model.Translation, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var translation model.Translation
	if err := db.Where("id = ?", id).First(&translation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get translation by id: %w", err)
	}
	return &translation, nil
}

// GetByThreadAndLanguages retrieves a translation by thread ID and language pair
func (r *TranslationRepo) GetByThreadAndLanguages(ctx context.Context, threadID, sourceLanguage, targetLanguage string) (*model.Translation, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var translation model.Translation
	if err := db.
		Where("thread_id = ? AND source_language = ? AND target_language = ?", threadID, sourceLanguage, targetLanguage).
		First(&translation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("get translation by thread and languages: %w", err)
	}
	return &translation, nil
}

// ListByThreadID retrieves all translations for a specific thread with pagination
func (r *TranslationRepo) ListByThreadID(ctx context.Context, threadID string, limit, offset int) ([]*model.Translation, int64, error) {
	db := sql.GetDBOrTx(ctx, r.db)
	var total int64
	if err := db.Model(&model.Translation{}).Where("thread_id = ?", threadID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count translations by thread: %w", err)
	}

	var translations []*model.Translation
	if err := db.
		Where("thread_id = ?", threadID).
		Limit(limit).
		Offset(offset).
		Find(&translations).Error; err != nil {
		return nil, 0, fmt.Errorf("get translations by thread: %w", err)
	}

	return translations, total, nil
}

// Update updates an existing translation
func (r *TranslationRepo) Update(ctx context.Context, translation *model.Translation) error {
	db := sql.GetDBOrTx(ctx, r.db)
	if err := db.Save(translation).Error; err != nil {
		return fmt.Errorf("update translation: %w", err)
	}
	return nil
}
