// Package repository provides a thin wrapper around the SQLC-generated Queries,
// keeping the service layer decoupled from raw database types.
package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	db "github.com/udayagiri/ainyx-backend/db/sqlc"
)

// ErrNotFound is returned when a requested user does not exist.
var ErrNotFound = errors.New("user not found")

// UserRepository is the interface the service layer depends on.
// Keeping this as an interface makes unit-testing easy via mocks.
type UserRepository interface {
	Create(ctx context.Context, name string, dob time.Time) (db.User, error)
	GetByID(ctx context.Context, id int32) (db.User, error)
	Update(ctx context.Context, id int32, name string, dob time.Time) (db.User, error)
	Delete(ctx context.Context, id int32) error
	List(ctx context.Context, limit, offset int32) ([]db.User, error)
	Count(ctx context.Context) (int64, error)
}

type userRepository struct {
	q *db.Queries
}

// New creates a new UserRepository backed by the given database connection.
func New(conn *sql.DB) UserRepository {
	return &userRepository{q: db.New(conn)}
}

func (r *userRepository) Create(ctx context.Context, name string, dob time.Time) (db.User, error) {
	return r.q.CreateUser(ctx, db.CreateUserParams{
		Name: name,
		Dob:  dob,
	})
}

func (r *userRepository) GetByID(ctx context.Context, id int32) (db.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return db.User{}, ErrNotFound
	}
	return u, err
}

func (r *userRepository) Update(ctx context.Context, id int32, name string, dob time.Time) (db.User, error) {
	u, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:   id,
		Name: name,
		Dob:  dob,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return db.User{}, ErrNotFound
	}
	return u, err
}

func (r *userRepository) Delete(ctx context.Context, id int32) error {
	// GetByID first so we can return ErrNotFound rather than a silent no-op.
	if _, err := r.GetByID(ctx, id); err != nil {
		return err
	}
	return r.q.DeleteUser(ctx, id)
}

func (r *userRepository) List(ctx context.Context, limit, offset int32) ([]db.User, error) {
	return r.q.ListUsers(ctx, db.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountUsers(ctx)
}
