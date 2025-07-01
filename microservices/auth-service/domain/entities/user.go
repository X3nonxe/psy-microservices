package entities

import "time"

type Role string

const (
	ClientRole       Role = "client"
	PsychologistRole Role = "psychologist"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email" validate:"email,required"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role" validate:"required"`
	CreatedAt    time.Time `json:"created_at"`
}

func (u *User) IsPsychologist() bool {
	return u.Role == PsychologistRole
}
