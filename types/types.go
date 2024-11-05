package types

import "time"

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByReferralCode(refCode string) (*User, error)
	SaveUser(user User) error
	GetUserById(userID int) (*User, error)
	SaveReferralCode(userID int, refCode string, expiry time.Time) error
	GetReferralsByReferrerID(referrerID int) ([]*User, error)
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	RefCode  string `json:"referral_code,omitempty"`
}

type User struct {
	ID           int        `json:id"`
	Email        string     `json:"email"`
	Password     string     `json:"password"`
	ReferrerID   *int       `json:"referrer_id"`
	ReferralCode *string    `json:"referral_code"`
	CodeExpiry   *time.Time `json:"code_expiry"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
