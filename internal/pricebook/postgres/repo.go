// Package postgres is the org_price_books repository.
package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"

	"github.com/Danish-13-Khan/backend-cabinet/internal/models/tables"
	"github.com/Danish-13-Khan/backend-cabinet/internal/pricebook"
	"github.com/Danish-13-Khan/backend-cabinet/internal/store"
	"github.com/Danish-13-Khan/backend-cabinet/internal/utils"
)

type Repo struct {
	db *gorm.DB
}

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Get(ctx context.Context, orgID string) (*pricebook.Book, error) {
	var m tables.OrgPriceBook
	err := store.GormDB(ctx, r.db).Where("org_id = ?", orgID).Take(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, utils.Generic(err)
	}
	return fromModel(&m), nil
}

// upsertSQL casts explicitly so the driver never sends the book as text.
const upsertSQL = `INSERT INTO org_price_books (org_id, book, updated_at, updated_by)
VALUES (?, ?::jsonb, ?, ?)
ON CONFLICT (org_id) DO UPDATE
SET book = EXCLUDED.book, updated_at = EXCLUDED.updated_at, updated_by = EXCLUDED.updated_by`

func (r *Repo) Put(ctx context.Context, b pricebook.Book) error {
	m := toModel(b)
	err := store.GormDB(ctx, r.db).Exec(upsertSQL, m.OrgID, m.Book, m.UpdatedAt, m.UpdatedBy).Error
	if err != nil {
		return utils.Generic(err)
	}
	return nil
}

func toModel(b pricebook.Book) tables.OrgPriceBook {
	return tables.OrgPriceBook{OrgID: b.OrgID, Book: string(b.Book), UpdatedAt: b.UpdatedAt.UTC(), UpdatedBy: b.UpdatedBy}
}

func fromModel(m *tables.OrgPriceBook) *pricebook.Book {
	return &pricebook.Book{OrgID: m.OrgID, Book: json.RawMessage(m.Book), UpdatedAt: m.UpdatedAt.UTC(), UpdatedBy: m.UpdatedBy}
}

var _ pricebook.Repo = (*Repo)(nil)
