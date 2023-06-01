package sqlc

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Note struct {
	ID  uuid.UUID
	Title  string
	Username  string
	Text  sql.NullString
	CreatedAt  time.Time
	UpdatedAt  time.Time
}