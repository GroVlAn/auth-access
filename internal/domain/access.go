package domain

import (
	"time"

	"github.com/GroVlAn/auth-base/ew"
)

type Role struct {
	ID          string    `json:"-" db:"id" valid:"require"`
	Name        string    `json:"name" db:"name" valid:"require"`
	Description string    `json:"description" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdateAt    time.Time `json:"update_at" db:"update_at"`
}

type Permission struct {
	ID          string    `json:"-" db:"id" valid:"require"`
	Name        string    `json:"name" db:"name" valid:"require"`
	Description string    `json:"description" db:"description"`
	IsDefault   bool      `json:"is_default" db:"is_default"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdateAt    time.Time `json:"update_at" db:"update_at"`
}

func (r Role) Validate() error {
	err := ew.NewErrValidation("validation role")

	if len(r.Name) == 0 {
		err.AddField("name", "name is required")
	}

	if err.IsEmpty() {
		return nil
	}

	return err
}

func (p Permission) Validate() error {
	err := ew.NewErrValidation("validation role")

	if len(p.Name) == 0 {
		err.AddField("name", "name is required")
	}

	if err.IsEmpty() {
		return nil
	}

	return err
}
