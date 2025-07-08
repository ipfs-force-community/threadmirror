package sqlrepo

import (
	"context"
	"errors"

	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"github.com/ipfs-force-community/threadmirror/pkg/errutil"
	"gorm.io/gorm"
)

type ThreadRepo struct{}

func NewThreadRepo() *ThreadRepo {
	return &ThreadRepo{}
}

func (r *ThreadRepo) GetThreadByID(ctx context.Context, id string) (*model.Thread, error) {
	db := sql.MustDBFromContext(ctx)
	var thread model.Thread
	err := db.Where("id = ?", id).First(&thread).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errutil.ErrNotFound
		}
		return nil, err
	}
	return &thread, nil
}

func (r *ThreadRepo) CreateThread(ctx context.Context, thread *model.Thread) error {
	db := sql.MustDBFromContext(ctx)
	return db.Create(thread).Error
}

func (r *ThreadRepo) GetTweetsByIDs(ctx context.Context, ids []string) (map[string]*model.Thread, error) {
	if len(ids) == 0 {
		return map[string]*model.Thread{}, nil
	}
	db := sql.MustDBFromContext(ctx)
	var tweets []model.Thread
	err := db.Where("id IN ?", ids).Find(&tweets).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*model.Thread, len(tweets))
	for i := range tweets {
		result[tweets[i].ID] = &tweets[i]
	}
	return result, nil
}
