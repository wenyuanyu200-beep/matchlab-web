package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type serviceRepository struct {
	usersFilter        UsersFilter
	activitiesFilter   ActivitiesFilter
	applicationsFilter ApplicationsFilter
	feedbacksPage      Page
	updatedTarget      uuid.UUID
	updatedRole        string
	updateCalls        int
}

func (r *serviceRepository) Stats(context.Context) (Stats, error) { return Stats{UsersCount: 2}, nil }
func (r *serviceRepository) Users(_ context.Context, filter UsersFilter) ([]AdminUser, error) {
	r.usersFilter = filter
	return nil, nil
}
func (r *serviceRepository) Activities(_ context.Context, filter ActivitiesFilter) ([]AdminActivity, error) {
	r.activitiesFilter = filter
	return nil, nil
}
func (r *serviceRepository) Applications(_ context.Context, filter ApplicationsFilter) ([]AdminApplication, error) {
	r.applicationsFilter = filter
	return nil, nil
}
func (r *serviceRepository) Feedbacks(_ context.Context, page Page) ([]Feedback, error) {
	r.feedbacksPage = page
	return nil, nil
}
func (r *serviceRepository) UpdateUserRole(_ context.Context, targetID uuid.UUID, role string) (*AdminUser, error) {
	r.updateCalls++
	r.updatedTarget = targetID
	r.updatedRole = role
	return &AdminUser{ID: targetID, Role: role}, nil
}

func TestServiceAppliesDefaultPaginationAndEmptyArrays(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)

	users, err := service.Users(context.Background(), UsersFilter{})
	if err != nil || users == nil || repository.usersFilter.Limit != 20 || repository.usersFilter.Offset != 0 {
		t.Fatalf("users=%v filter=%#v err=%v", users, repository.usersFilter, err)
	}
	feedbacks, err := service.Feedbacks(context.Background(), Page{})
	if err != nil || feedbacks == nil || repository.feedbacksPage.Limit != 20 {
		t.Fatalf("feedbacks=%v page=%#v err=%v", feedbacks, repository.feedbacksPage, err)
	}
}

func TestServiceValidatesFiltersAndPagination(t *testing.T) {
	service := NewService(&serviceRepository{})
	tests := []struct {
		name string
		call func() error
	}{
		{name: "invalid user role", call: func() error { _, err := service.Users(context.Background(), UsersFilter{Role: "owner"}); return err }},
		{name: "invalid activity status", call: func() error {
			_, err := service.Activities(context.Background(), ActivitiesFilter{Status: "draft"})
			return err
		}},
		{name: "invalid application status", call: func() error {
			_, err := service.Applications(context.Background(), ApplicationsFilter{Status: "accepted"})
			return err
		}},
		{name: "limit too high", call: func() error { _, err := service.Feedbacks(context.Background(), Page{Limit: 101}); return err }},
		{name: "negative limit", call: func() error { _, err := service.Feedbacks(context.Background(), Page{Limit: -1}); return err }},
		{name: "negative offset", call: func() error { _, err := service.Feedbacks(context.Background(), Page{Offset: -1}); return err }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.call(); !errors.Is(err, ErrInvalidFilter) {
				t.Fatalf("err=%v, want ErrInvalidFilter", err)
			}
		})
	}
}

func TestServiceAllowsValidFilters(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)
	if _, err := service.Users(context.Background(), UsersFilter{Role: "admin", Page: Page{Limit: 100, Offset: 3}}); err != nil {
		t.Fatalf("users: %v", err)
	}
	if repository.usersFilter.Role != "admin" || repository.usersFilter.Limit != 100 || repository.usersFilter.Offset != 3 {
		t.Fatalf("filter=%#v", repository.usersFilter)
	}
	if _, err := service.Activities(context.Background(), ActivitiesFilter{Status: "recruiting"}); err != nil {
		t.Fatalf("activities: %v", err)
	}
	if _, err := service.Applications(context.Background(), ApplicationsFilter{Status: "approved"}); err != nil {
		t.Fatalf("applications: %v", err)
	}
}

func TestServiceRejectsInvalidRoleAndSelfDemotion(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)
	requesterID := uuid.New()
	if _, err := service.UpdateUserRole(context.Background(), requesterID, uuid.New(), "owner"); !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("invalid role err=%v", err)
	}
	if _, err := service.UpdateUserRole(context.Background(), requesterID, requesterID, "user"); !errors.Is(err, ErrSelfDemotion) {
		t.Fatalf("self demotion err=%v", err)
	}
	if repository.updateCalls != 0 {
		t.Fatalf("unexpected update calls=%d", repository.updateCalls)
	}
}

func TestServiceUpdatesOtherUserRole(t *testing.T) {
	repository := &serviceRepository{}
	service := NewService(repository)
	targetID := uuid.New()
	updated, err := service.UpdateUserRole(context.Background(), uuid.New(), targetID, "admin")
	if err != nil || updated.ID != targetID || repository.updatedRole != "admin" {
		t.Fatalf("updated=%#v repository=%#v err=%v", updated, repository, err)
	}
}

var _ Repository = (*serviceRepository)(nil)
