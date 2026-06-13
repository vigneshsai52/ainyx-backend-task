// Package service contains business logic that sits between handlers and the
// repository.  Age calculation lives here so it is easy to unit-test in
// isolation.
package service

import (
	"context"
	"time"

	"github.com/udayagiri/ainyx-backend/internal/logger"
	"github.com/udayagiri/ainyx-backend/internal/models"
	"github.com/udayagiri/ainyx-backend/internal/repository"
)

const dobLayout = "2006-01-02"

// UserService defines the operations available to the handler layer.
type UserService interface {
	Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error)
	GetByID(ctx context.Context, id int32) (models.UserWithAgeResponse, error)
	Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, page, pageSize int) (models.PaginatedUsersResponse, error)
}

type userService struct {
	repo repository.UserRepository
}

// New creates a new UserService.
func New(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

// ── Create ────────────────────────────────────────────────────────────────────

func (s *userService) Create(ctx context.Context, req models.CreateUserRequest) (models.UserResponse, error) {
	log := logger.Get()

	dob, err := time.Parse(dobLayout, req.Dob)
	if err != nil {
		return models.UserResponse{}, err
	}

	u, err := s.repo.Create(ctx, req.Name, dob)
	if err != nil {
		log.Errorw("failed to create user", "error", err)
		return models.UserResponse{}, err
	}

	log.Infow("user created", "id", u.ID, "name", u.Name)
	return models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		Dob:  u.Dob.Format(dobLayout),
	}, nil
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func (s *userService) GetByID(ctx context.Context, id int32) (models.UserWithAgeResponse, error) {
	log := logger.Get()

	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		log.Warnw("user not found", "id", id)
		return models.UserWithAgeResponse{}, err
	}

	age := CalculateAge(u.Dob, time.Now())
	log.Infow("user fetched", "id", u.ID)

	return models.UserWithAgeResponse{
		ID:   u.ID,
		Name: u.Name,
		Dob:  u.Dob.Format(dobLayout),
		Age:  age,
	}, nil
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *userService) Update(ctx context.Context, id int32, req models.UpdateUserRequest) (models.UserResponse, error) {
	log := logger.Get()

	dob, err := time.Parse(dobLayout, req.Dob)
	if err != nil {
		return models.UserResponse{}, err
	}

	u, err := s.repo.Update(ctx, id, req.Name, dob)
	if err != nil {
		log.Warnw("update failed", "id", id, "error", err)
		return models.UserResponse{}, err
	}

	log.Infow("user updated", "id", u.ID)
	return models.UserResponse{
		ID:   u.ID,
		Name: u.Name,
		Dob:  u.Dob.Format(dobLayout),
	}, nil
}

// ── Delete ────────────────────────────────────────────────────────────────────

func (s *userService) Delete(ctx context.Context, id int32) error {
	log := logger.Get()

	if err := s.repo.Delete(ctx, id); err != nil {
		log.Warnw("delete failed", "id", id, "error", err)
		return err
	}

	log.Infow("user deleted", "id", id)
	return nil
}

// ── List ──────────────────────────────────────────────────────────────────────

func (s *userService) List(ctx context.Context, page, pageSize int) (models.PaginatedUsersResponse, error) {
	log := logger.Get()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, err := s.repo.List(ctx, int32(pageSize), int32(offset))
	if err != nil {
		log.Errorw("failed to list users", "error", err)
		return models.PaginatedUsersResponse{}, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		log.Errorw("failed to count users", "error", err)
		return models.PaginatedUsersResponse{}, err
	}

	now := time.Now()
	data := make([]models.UserWithAgeResponse, len(users))
	for i, u := range users {
		data[i] = models.UserWithAgeResponse{
			ID:   u.ID,
			Name: u.Name,
			Dob:  u.Dob.Format(dobLayout),
			Age:  CalculateAge(u.Dob, now),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	log.Infow("users listed", "page", page, "pageSize", pageSize, "total", total)
	return models.PaginatedUsersResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// ── CalculateAge ──────────────────────────────────────────────────────────────

// CalculateAge computes the number of complete years between dob and now.
// It is exported so it can be unit-tested directly.
//
// The algorithm:
//  1. Subtract birth year from current year to get a naive age.
//  2. If the birthday has not occurred yet this year (month/day comparison),
//     subtract one.
func CalculateAge(dob, now time.Time) int {
	years := now.Year() - dob.Year()

	// Has the birthday already passed this calendar year?
	birthdayThisYear := time.Date(now.Year(), dob.Month(), dob.Day(), 0, 0, 0, 0, now.Location())
	if now.Before(birthdayThisYear) {
		years--
	}

	if years < 0 {
		return 0
	}
	return years
}
