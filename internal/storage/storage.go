package storage

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrReferralCodeNotFound = errors.New("referral code not found")
	ErrDuplicateEmail       = errors.New("email already exists")
	ErrDuplicateReferral    = errors.New("referral code already exists")
	ErrDatabase             = errors.New("database error")
)
