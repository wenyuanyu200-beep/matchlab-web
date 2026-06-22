package circle

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidName        = errors.New("name must be 2 to 40 characters")
	ErrInvalidDescription = errors.New("description must not exceed 300 characters")
	ErrInvalidCategory    = errors.New("invalid category")
	ErrInvalidStatus      = errors.New("status must be pending, approved, or rejected")
)

var allowedCategories = map[string]bool{
	"study": true, "sports": true, "interest": true, "career": true, "other": true, "general": true,
}

type Service interface {
	ListApproved(ctx context.Context) ([]Circle, error)
	GetApproved(ctx context.Context, id uuid.UUID) (*Circle, error)
	Members(ctx context.Context, id uuid.UUID) ([]MemberSummary, error)
	Create(ctx context.Context, userID uuid.UUID, name, description, category string) (*Circle, *Channel, error)
	Join(ctx context.Context, circleID, userID uuid.UUID) error
	Mine(ctx context.Context, userID uuid.UUID) ([]Circle, error)
	Channels(ctx context.Context, circleID, userID uuid.UUID) ([]Channel, error)
	Messages(ctx context.Context, circleID, channelID, userID uuid.UUID) ([]Message, error)
	PostMessage(ctx context.Context, circleID, channelID, senderID uuid.UUID, content string) (*Message, error)
	AdminList(ctx context.Context, filter AdminFilter) ([]Circle, error)
	Moderate(ctx context.Context, id uuid.UUID, status string) (*Circle, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) ListApproved(ctx context.Context) ([]Circle, error) {
	rows, err := s.repository.ListApproved(ctx)
	if rows == nil && err == nil {
		rows = []Circle{}
	}
	return rows, err
}

func (s *service) GetApproved(ctx context.Context, id uuid.UUID) (*Circle, error) {
	return s.repository.GetApproved(ctx, id)
}

func (s *service) Members(ctx context.Context, id uuid.UUID) ([]MemberSummary, error) {
	rows, err := s.repository.Members(ctx, id)
	if rows == nil && err == nil {
		rows = []MemberSummary{}
	}
	return rows, err
}

func (s *service) Create(ctx context.Context, userID uuid.UUID, name, description, category string) (*Circle, *Channel, error) {
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	category = strings.TrimSpace(category)

	nameLen := utf8.RuneCountInString(name)
	if nameLen < 2 || nameLen > 40 {
		return nil, nil, ErrInvalidName
	}
	if utf8.RuneCountInString(description) > 300 {
		return nil, nil, ErrInvalidDescription
	}
	if category == "" {
		category = "general"
	}
	if !allowedCategories[category] {
		return nil, nil, ErrInvalidCategory
	}

	row := Circle{
		CreatorID:    userID,
		Name:        name,
		Description: description,
		Category:    category,
		Status:      "pending",
		MemberCount: 1,
	}
	channel, err := s.repository.Create(ctx, &row)
	return &row, channel, err
}

func (s *service) Join(ctx context.Context, circleID, userID uuid.UUID) error {
	return s.repository.Join(ctx, circleID, userID)
}

func (s *service) Mine(ctx context.Context, userID uuid.UUID) ([]Circle, error) {
	rows, err := s.repository.Mine(ctx, userID)
	if rows == nil && err == nil {
		rows = []Circle{}
	}
	return rows, err
}

func (s *service) Channels(ctx context.Context, circleID, userID uuid.UUID) ([]Channel, error) {
	rows, err := s.repository.Channels(ctx, circleID, userID)
	if rows == nil && err == nil {
		rows = []Channel{}
	}
	return rows, err
}

func (s *service) Messages(ctx context.Context, circleID, channelID, userID uuid.UUID) ([]Message, error) {
	rows, err := s.repository.Messages(ctx, circleID, channelID, userID)
	if rows == nil && err == nil {
		rows = []Message{}
	}
	return rows, err
}

func (s *service) PostMessage(ctx context.Context, circleID, channelID, senderID uuid.UUID, content string) (*Message, error) {
	content = strings.TrimSpace(content)
	contentLen := utf8.RuneCountInString(content)
	if contentLen < 1 || contentLen > 1000 {
		return nil, errors.New("content must contain 1 to 1000 characters")
	}
	row := Message{
		CircleID:  circleID,
		ChannelID: channelID,
		SenderID:  senderID,
		Content:   content,
	}
	if err := s.repository.PostMessage(ctx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *service) AdminList(ctx context.Context, filter AdminFilter) ([]Circle, error) {
	filter.Status = strings.TrimSpace(strings.ToLower(filter.Status))
	if filter.Status != "" && filter.Status != "pending" && filter.Status != "approved" && filter.Status != "rejected" {
		return nil, ErrInvalidStatus
	}
	rows, err := s.repository.AdminList(ctx, filter)
	if rows == nil && err == nil {
		rows = []Circle{}
	}
	return rows, err
}

func (s *service) Moderate(ctx context.Context, id uuid.UUID, status string) (*Circle, error) {
	status = strings.TrimSpace(strings.ToLower(status))
	if status != "approved" && status != "rejected" {
		return nil, ErrInvalidStatus
	}
	return s.repository.Moderate(ctx, id, status)
}
