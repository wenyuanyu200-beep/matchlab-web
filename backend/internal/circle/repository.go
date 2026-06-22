package circle

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrUnavailable = errors.New("circle repository unavailable")
	ErrNotFound = errors.New("circle not found")
	ErrForbidden = errors.New("circle membership required")
	ErrAlreadyMember = errors.New("already a circle member")
	ErrInvalidState = errors.New("invalid circle state")
)

type Repository interface {
	ListApproved(context.Context) ([]Circle, error)
	GetApproved(context.Context, uuid.UUID) (*Circle, error)
	Members(context.Context, uuid.UUID) ([]MemberSummary, error)
	Create(context.Context, *Circle) (*Channel, error)
	Join(context.Context, uuid.UUID, uuid.UUID) error
	Mine(context.Context, uuid.UUID) ([]Circle, error)
	Channels(context.Context, uuid.UUID, uuid.UUID) ([]Channel, error)
	Messages(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) ([]Message, error)
	PostMessage(context.Context, *Message) error
	AdminList(context.Context, AdminFilter) ([]Circle, error)
	Moderate(context.Context, uuid.UUID, string) (*Circle, error)
}

type GormRepository struct{ db *gorm.DB }
func NewGormRepository(db *gorm.DB) *GormRepository { return &GormRepository{db: db} }

func (r *GormRepository) ListApproved(ctx context.Context) ([]Circle, error) {
	if r.db == nil { return nil, ErrUnavailable }
	var rows []Circle
	err := r.db.WithContext(ctx).Where("status = ?", "approved").Order("created_at DESC").Find(&rows).Error
	return rows, mapError(err)
}
func (r *GormRepository) GetApproved(ctx context.Context, id uuid.UUID) (*Circle, error) {
	if r.db == nil { return nil, ErrUnavailable }
	var row Circle
	if err := r.db.WithContext(ctx).Where("id = ? AND status = ?", id, "approved").First(&row).Error; err != nil { return nil, mapError(err) }
	return &row, nil
}
func (r *GormRepository) Members(ctx context.Context, id uuid.UUID) ([]MemberSummary, error) {
	if r.db == nil { return nil, ErrUnavailable }
	if _, err := r.GetApproved(ctx, id); err != nil { return nil, err }
	var rows []MemberSummary
	err := r.db.WithContext(ctx).Table("circle_members cm").Select("u.id, u.nickname, u.school, cm.role, cm.joined_at").Joins("JOIN users u ON u.id = cm.user_id").Where("cm.circle_id = ?", id).Order("cm.joined_at").Scan(&rows).Error
	return rows, mapError(err)
}
func (r *GormRepository) Create(ctx context.Context, model *Circle) (*Channel, error) {
	if r.db == nil { return nil, ErrUnavailable }
	var channel Channel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(model).Error; err != nil { return mapError(err) }
		if err := tx.Exec("INSERT INTO circle_members (circle_id,user_id,role) VALUES (?,?,?)", model.ID, model.CreatorID, "owner").Error; err != nil { return mapError(err) }
		channel = Channel{CircleID: model.ID, Name: "general"}
		return mapError(tx.Create(&channel).Error)
	})
	return &channel, err
}
func (r *GormRepository) Join(ctx context.Context, circleID, userID uuid.UUID) error {
	if r.db == nil { return ErrUnavailable }
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row Circle
		if err := tx.Clauses(clause.Locking{Strength:"UPDATE"}).First(&row, "id = ?", circleID).Error; err != nil { return mapError(err) }
		if row.Status != "approved" { return ErrInvalidState }
		result := tx.Exec("INSERT INTO circle_members (circle_id,user_id,role) VALUES (?,?,?) ON CONFLICT DO NOTHING", circleID, userID, "member")
		if result.Error != nil { return mapError(result.Error) }
		if result.RowsAffected == 0 { return ErrAlreadyMember }
		return mapError(tx.Model(&Circle{}).Where("id = ?", circleID).UpdateColumn("member_count", gorm.Expr("member_count + 1")).Error)
	})
}
func (r *GormRepository) Mine(ctx context.Context, userID uuid.UUID) ([]Circle, error) {
	if r.db == nil { return nil, ErrUnavailable }
	var rows []Circle
	err := r.db.WithContext(ctx).Table("circles c").Select("c.*").Joins("JOIN circle_members cm ON cm.circle_id=c.id").Where("cm.user_id = ?", userID).Order("c.created_at DESC").Scan(&rows).Error
	return rows, mapError(err)
}
func (r *GormRepository) requireMember(ctx context.Context, circleID, userID uuid.UUID) error {
	var count int64
	if err := r.db.WithContext(ctx).Table("circle_members").Where("circle_id=? AND user_id=?", circleID, userID).Count(&count).Error; err != nil { return mapError(err) }
	if count == 0 { return ErrForbidden }; return nil
}
func (r *GormRepository) Channels(ctx context.Context, circleID, userID uuid.UUID) ([]Channel, error) {
	if r.db == nil { return nil, ErrUnavailable }; if err := r.requireMember(ctx,circleID,userID); err != nil{return nil,err}
	var rows []Channel; err:=r.db.WithContext(ctx).Where("circle_id=?",circleID).Order("created_at").Find(&rows).Error; return rows,mapError(err)
}
func (r *GormRepository) Messages(ctx context.Context, circleID, channelID, userID uuid.UUID) ([]Message, error) {
	if r.db == nil { return nil, ErrUnavailable }; if err:=r.requireMember(ctx,circleID,userID);err!=nil{return nil,err}
	var count int64; if err:=r.db.WithContext(ctx).Table("circle_channels").Where("id=? AND circle_id=?",channelID,circleID).Count(&count).Error;err!=nil{return nil,mapError(err)};if count==0{return nil,ErrNotFound}
	var rows []Message; err:=r.db.WithContext(ctx).Table("circle_messages m").Select("m.*,u.nickname AS sender_nickname").Joins("JOIN users u ON u.id=m.sender_id").Where("m.circle_id=? AND m.channel_id=?",circleID,channelID).Order("m.created_at").Limit(100).Scan(&rows).Error;return rows,mapError(err)
}
func (r *GormRepository) PostMessage(ctx context.Context, m *Message) error {
	if r.db == nil{return ErrUnavailable};if err:=r.requireMember(ctx,m.CircleID,m.SenderID);err!=nil{return err}
	var count int64;if err:=r.db.WithContext(ctx).Table("circle_channels").Where("id=? AND circle_id=?",m.ChannelID,m.CircleID).Count(&count).Error;err!=nil{return mapError(err)};if count==0{return ErrNotFound};return mapError(r.db.WithContext(ctx).Create(m).Error)
}
func (r *GormRepository) AdminList(ctx context.Context, f AdminFilter)([]Circle,error){if r.db==nil{return nil,ErrUnavailable};var rows []Circle;q:=r.db.WithContext(ctx);if f.Status!=""{q=q.Where("status=?",f.Status)};err:=q.Order("created_at DESC").Find(&rows).Error;return rows,mapError(err)}
func (r *GormRepository) Moderate(ctx context.Context,id uuid.UUID,status string)(*Circle,error){if r.db==nil{return nil,ErrUnavailable};var row Circle;if err:=r.db.WithContext(ctx).First(&row,"id=?",id).Error;err!=nil{return nil,mapError(err)};if row.Status!=status{row.Status=status;if err:=r.db.WithContext(ctx).Save(&row).Error;err!=nil{return nil,mapError(err)}};return &row,nil}
func mapError(err error) error { if err==nil{return nil};if errors.Is(err,gorm.ErrRecordNotFound){return ErrNotFound};var pg *pgconn.PgError;if errors.As(err,&pg)&&pg.Code=="23505"{return ErrAlreadyMember};return fmt.Errorf("circle repository: %w",err) }
