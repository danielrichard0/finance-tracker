package model

import (
	"database/sql"
	"time"
)

type User struct {
	UserID    string       `json:"user_id"`
	Email     string       `json:"email"`
	Phone     string       `json:"phone"`
	Username  string       `json:"username"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt sql.NullTime `json:"updated_at"`
}
