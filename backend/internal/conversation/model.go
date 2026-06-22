package conversation

import("time";"github.com/google/uuid")
type Conversation struct{ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`;Kind string `json:"kind"`;DirectKey string `gorm:"column:direct_key" json:"-"`;LastMessageAt *time.Time `gorm:"column:last_message_at" json:"last_message_at"`;CreatedAt time.Time `json:"created_at"`;UpdatedAt time.Time `json:"updated_at"`}
type UserSummary struct{ID uuid.UUID `json:"id"`;Nickname string `json:"nickname"`;School string `json:"school"`}
type ConversationSummary struct{Conversation;Peer UserSummary `json:"peer"`;UnreadCount int64 `json:"unread_count"`}
type Message struct{ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`;ConversationID uuid.UUID `gorm:"column:conversation_id" json:"conversation_id"`;SenderID uuid.UUID `gorm:"column:sender_id" json:"sender_id"`;Content string `json:"content"`;CreatedAt time.Time `json:"created_at"`;Sender UserSummary `gorm:"-" json:"sender"`}
