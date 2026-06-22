package conversation

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var (
	ErrInvalidContent = errors.New("content must contain 1 to 1000 characters")
)

type Service interface {
	Direct(ctx context.Context, me, peer uuid.UUID) (*Conversation, error)
	List(ctx context.Context, uid uuid.UUID) ([]ConversationSummary, error)
	Messages(ctx context.Context, cid, uid uuid.UUID) ([]Message, error)
	Post(ctx context.Context, cid, senderID uuid.UUID, content string) (*Message, error)
	Read(ctx context.Context, cid, uid uuid.UUID) error
	Unread(ctx context.Context, uid uuid.UUID) (int64, error)
}

type service struct {
	repository Repository
}

func NewService(repository Repository) Service {
	return &service{repository: repository}
}

func (s *service) Direct(ctx context.Context, me, peer uuid.UUID) (*Conversation, error) {
	return s.repository.Direct(ctx, me, peer)
}

func (s *service) List(ctx context.Context, uid uuid.UUID) ([]ConversationSummary, error) {
	rows, err := s.repository.List(ctx, uid)
	if rows == nil && err == nil {
		rows = []ConversationSummary{}
	}
	return rows, err
}

func (s *service) Messages(ctx context.Context, cid, uid uuid.UUID) ([]Message, error) {
	rows, err := s.repository.Messages(ctx, cid, uid)
	if rows == nil && err == nil {
		rows = []Message{}
	}
	return rows, err
}

func (s *service) Post(ctx context.Context, cid, senderID uuid.UUID, content string) (*Message, error) {
	content = strings.TrimSpace(content)
	contentLen := utf8.RuneCountInString(content)
	if contentLen < 1 || contentLen > 1000 {
		return nil, ErrInvalidContent
	}
	row := Message{
		ConversationID: cid,
		SenderID:       senderID,
		Content:        content,
	}
	if err := s.repository.Post(ctx, &row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *service) Read(ctx context.Context, cid, uid uuid.UUID) error {
	return s.repository.Read(ctx, cid, uid)
}

func (s *service) Unread(ctx context.Context, uid uuid.UUID) (int64, error) {
	return s.repository.Unread(ctx, uid)
}
