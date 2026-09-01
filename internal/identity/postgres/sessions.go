package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Danish-13-Khan/backend-cabinet/internal/identity"
)

type SessionRepo struct {
	db *gorm.DB
}

func NewSessionRepo(db *gorm.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) GetByID(ctx context.Context, id string) (*identity.Session, error) {
	var m SessionModel
	if err := use(ctx, r.db).Where("public_id = ?", id).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return sessionFromModel(&m), nil
}

func (r *SessionRepo) GetByRefreshHash(ctx context.Context, hash string) (*identity.Session, error) {
	var m SessionModel
	if err := use(ctx, r.db).Where("refresh_token_hash = ?", hash).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return sessionFromModel(&m), nil
}

func (r *SessionRepo) GetByRefreshHashForUpdate(ctx context.Context, hash string) (*identity.Session, error) {
	var m SessionModel
	if err := use(ctx, r.db).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("refresh_token_hash = ?", hash).Take(&m).Error; err != nil {
		return nil, mapGorm(err)
	}
	return sessionFromModel(&m), nil
}

func (r *SessionRepo) Insert(ctx context.Context, s *identity.Session) error {
	m := sessionToModel(s)
	if err := use(ctx, r.db).Create(&m).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *SessionRepo) RevokeAndReplace(ctx context.Context, oldID, newID string, at time.Time) error {
	res := use(ctx, r.db).Model(&SessionModel{}).Where("public_id = ?", oldID).Updates(map[string]any{
		"revoked_at":             at,
		"replaced_by_session_id": newID,
		"updated_at":             at,
	})
	if res.Error != nil {
		return mapGorm(res.Error)
	}
	if res.RowsAffected == 0 {
		return identity.ErrNotFound
	}
	return nil
}

func (r *SessionRepo) MarkReuseAndRevokeFamily(ctx context.Context, sessionID, familyID string, at time.Time) error {
	if err := use(ctx, r.db).Model(&SessionModel{}).Where("public_id = ?", sessionID).Updates(map[string]any{
		"reuse_detected_at": at,
		"revoked_at":        gorm.Expr("COALESCE(revoked_at, ?)", at),
		"updated_at":        at,
	}).Error; err != nil {
		return mapGorm(err)
	}
	if err := use(ctx, r.db).Model(&SessionModel{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Updates(map[string]any{
			"revoked_at": at,
			"updated_at": at,
		}).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *SessionRepo) Revoke(ctx context.Context, id string, at time.Time) error {
	if err := use(ctx, r.db).Model(&SessionModel{}).
		Where("public_id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{
			"revoked_at": at,
			"updated_at": at,
		}).Error; err != nil {
		return mapGorm(err)
	}
	return nil
}

func (r *SessionRepo) RevokeAllForUser(ctx context.Context, userID string, at time.Time) (int, error) {
	res := use(ctx, r.db).Model(&SessionModel{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Updates(map[string]any{
			"revoked_at": at,
			"updated_at": at,
		})
	if res.Error != nil {
		return 0, mapGorm(res.Error)
	}
	return int(res.RowsAffected), nil
}

var _ identity.SessionRepo = (*SessionRepo)(nil)
