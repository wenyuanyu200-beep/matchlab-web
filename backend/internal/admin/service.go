package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrInvalidFilter = errors.New("invalid admin filter")
	ErrInvalidRole   = errors.New("role must be user or admin")
	ErrSelfDemotion  = errors.New("cannot demote the current administrator")
)

type Repository interface {
	Stats(ctx context.Context) (Stats, error)
	Users(ctx context.Context, filter UsersFilter) ([]AdminUser, error)
	Activities(ctx context.Context, filter ActivitiesFilter) ([]AdminActivity, error)
	Applications(ctx context.Context, filter ApplicationsFilter) ([]AdminApplication, error)
	Feedbacks(ctx context.Context, page Page) ([]Feedback, error)
	UpdateUserRole(ctx context.Context, targetID uuid.UUID, role string) (*AdminUser, error)
}

type Service interface {
	Stats(ctx context.Context) (Stats, error)
	Users(ctx context.Context, filter UsersFilter) ([]AdminUser, error)
	Activities(ctx context.Context, filter ActivitiesFilter) ([]AdminActivity, error)
	Applications(ctx context.Context, filter ApplicationsFilter) ([]AdminApplication, error)
	Feedbacks(ctx context.Context, page Page) ([]Feedback, error)
	UpdateUserRole(ctx context.Context, requesterID, targetID uuid.UUID, role string) (*AdminUser, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service { return &service{repository: repository} }

func (s *service) Stats(ctx context.Context) (Stats, error) {
	return s.repository.Stats(ctx)
}

func (s *service) Users(ctx context.Context, filter UsersFilter) ([]AdminUser, error) {
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	filter.Role = strings.ToLower(strings.TrimSpace(filter.Role))
	if filter.Role != "" && filter.Role != "user" && filter.Role != "admin" {
		return nil, ErrInvalidFilter
	}
	page, err := normalizePage(filter.Page)
	if err != nil {
		return nil, err
	}
	filter.Page = page
	users, err := s.repository.Users(ctx, filter)
	if users == nil && err == nil {
		users = []AdminUser{}
	}
	return users, err
}

func (s *service) Activities(ctx context.Context, filter ActivitiesFilter) ([]AdminActivity, error) {
	filter.Keyword = strings.TrimSpace(filter.Keyword)
	filter.Type = strings.TrimSpace(filter.Type)
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if !allowed(filter.Status, "recruiting", "full", "closed") {
		return nil, ErrInvalidFilter
	}
	page, err := normalizePage(filter.Page)
	if err != nil {
		return nil, err
	}
	filter.Page = page
	activities, err := s.repository.Activities(ctx, filter)
	if activities == nil && err == nil {
		activities = []AdminActivity{}
	}
	return activities, err
}

func (s *service) Applications(ctx context.Context, filter ApplicationsFilter) ([]AdminApplication, error) {
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	if !allowed(filter.Status, "pending", "approved", "rejected", "cancelled") {
		return nil, ErrInvalidFilter
	}
	page, err := normalizePage(filter.Page)
	if err != nil {
		return nil, err
	}
	filter.Page = page
	applications, err := s.repository.Applications(ctx, filter)
	if applications == nil && err == nil {
		applications = []AdminApplication{}
	}
	return applications, err
}

func (s *service) Feedbacks(ctx context.Context, page Page) ([]Feedback, error) {
	page, err := normalizePage(page)
	if err != nil {
		return nil, err
	}
	feedbacks, err := s.repository.Feedbacks(ctx, page)
	if feedbacks == nil && err == nil {
		feedbacks = []Feedback{}
	}
	return feedbacks, err
}

func (s *service) UpdateUserRole(ctx context.Context, requesterID, targetID uuid.UUID, role string) (*AdminUser, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "user" && role != "admin" {
		return nil, ErrInvalidRole
	}
	if requesterID == targetID && role == "user" {
		return nil, ErrSelfDemotion
	}
	return s.repository.UpdateUserRole(ctx, targetID, role)
}

func normalizePage(page Page) (Page, error) {
	if page.Limit == 0 {
		page.Limit = 20
	}
	if page.Limit < 1 || page.Limit > 100 || page.Offset < 0 {
		return Page{}, ErrInvalidFilter
	}
	return page, nil
}

func allowed(value string, options ...string) bool {
	if value == "" {
		return true
	}
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
