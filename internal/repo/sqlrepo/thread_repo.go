package sqlrepo

import (
	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
)

// ThreadRepo 提供 Thread 的基本数据库操作
type ThreadRepo struct {
	db *sql.DB
}

func NewThreadRepo(db *sql.DB) *ThreadRepo {
	return &ThreadRepo{db: db}
}

func (r *ThreadRepo) GetThreadByID(id string) (*model.Thread, error) {
	var thread model.Thread
	err := r.db.Where("id = ?", id).First(&thread).Error
	if err != nil {
		return nil, err
	}
	return &thread, nil
}

func (r *ThreadRepo) CreateThread(thread *model.Thread) error {
	return r.db.Create(thread).Error
}

func (r *ThreadRepo) GetThreadsByIDs(ids []string) (map[string]*model.Thread, error) {
	if len(ids) == 0 {
		return map[string]*model.Thread{}, nil
	}
	var threads []model.Thread
	err := r.db.Where("id IN ?", ids).Find(&threads).Error
	if err != nil {
		return nil, err
	}
	result := make(map[string]*model.Thread, len(threads))
	for i := range threads {
		result[threads[i].ID] = &threads[i]
	}
	return result, nil
}
