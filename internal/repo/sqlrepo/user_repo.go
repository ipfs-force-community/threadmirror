package sqlrepo

import (
	"github.com/ipfs-force-community/threadmirror/internal/model"
	"github.com/ipfs-force-community/threadmirror/pkg/database/sql"
	"gorm.io/datatypes"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetUserByID retrieves a user by ID
func (r *UserRepo) GetUserByID(id datatypes.UUID) (*model.UserProfile, error) {
	var user model.UserProfile
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByDisplayID retrieves a user by display ID
func (r *UserRepo) GetUserByDisplayID(displayID string) (*model.UserProfile, error) {
	var user model.UserProfile
	if err := r.db.Where("display_id = ?", displayID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user
func (r *UserRepo) CreateUser(user *model.UserProfile) error {
	return r.db.Create(user).Error
}

// UpdateUser updates an existing user
func (r *UserRepo) UpdateUser(user *model.UserProfile) error {
	return r.db.Save(user).Error
}

// DeleteUser soft deletes a user
func (r *UserRepo) DeleteUser(id datatypes.UUID) error {
	return r.db.Where("id = ?", id).Delete(&model.UserProfile{}).Error
}

// SearchUsers searches users by nickname with pagination
func (r *UserRepo) SearchUsers(
	query string,
	limit, offset int,
) ([]model.UserProfile, int64, error) {
	var users []model.UserProfile
	var total int64

	// Count total matching users
	if err := r.db.Model(&model.UserProfile{}).
		Where("nickname LIKE ?", "%"+query+"%").
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get matching users with pagination
	if err := r.db.Where("nickname LIKE ?", "%"+query+"%").
		Order("nickname").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
