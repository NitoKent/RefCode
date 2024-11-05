package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"RefCode.com/m/internal/storage"
	"RefCode.com/m/types"
	_ "github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}
type scanRowInterface interface {
	Scan(dest ...interface{}) error
}

func NewStore(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to open database: %w", op, err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		referrer_id INTEGER NULL,
		referral_code TEXT,
		code_expiry TIMESTAMP,
		FOREIGN KEY (referrer_id) REFERENCES users(id)
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: unable to create table: %w", op, err)
	}

	return &Storage{db: db}, nil
}
func (s *Storage) GetUserByReferralCode(refCode string) (*types.User, error) {
	row := s.db.QueryRow("SELECT id, email, referrer_id, referral_code FROM users WHERE referral_code = ?", refCode)
	user, err := scanRowIntoUser(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no user found with referral code %s", refCode)
	} else if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Storage) GetUserByEmail(email string) (*types.User, error) {
	//
	row := s.db.QueryRow("SELECT id, email, password, referrer_id, referral_code FROM users WHERE email = ?", email)

	u := new(types.User)
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.ReferrerID, &u.ReferralCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}

	return u, nil
}

func scanRowIntoUser(scanner scanRowInterface) (*types.User, error) {
	user := new(types.User)
	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.ReferrerID,
		&user.ReferralCode,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Storage) SaveUser(user types.User) error {
	const op = "storage.sqlite.SaveUser"

	var referrerID interface{}
	if user.ReferrerID != nil {
		referrerID = *user.ReferrerID
	} else {
		referrerID = nil
	}

	_, err := s.db.Exec(
		"INSERT INTO users (email, password, referrer_id) VALUES (?, ?, ?)",
		user.Email, user.Password, referrerID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert user: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserById(id int) (*types.User, error) {
	row := s.db.QueryRow("SELECT id, email, password, referrer_id, referral_code FROM users WHERE id = ?", id)

	u := new(types.User)
	if err := row.Scan(&u.ID, &u.Email, &u.Password, &u.ReferrerID, &u.ReferralCode); err != nil {
		if err == sql.ErrNoRows {
			return nil, storage.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (s *Storage) SaveReferralCode(userID int, refCode string, expiry time.Time) error {
	const op = "storage.sqlite.SaveReferralCode"
	_, err := s.db.Exec(
		"UPDATE users SET referral_code = ?, code_expiry = ? WHERE id = ?",
		refCode, expiry, userID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to save referral code: %w", op, err)
	}
	return nil
}

func (s *Storage) GetReferralsByReferrerID(referrerID int) ([]*types.User, error) {
	rows, err := s.db.Query("SELECT id, email, referrer_id, referral_code FROM users WHERE referrer_id = ?", referrerID)
	if err != nil {
		return nil, fmt.Errorf("could not get referrals: %w", err)
	}
	defer rows.Close()

	var referrals []*types.User
	for rows.Next() {
		user, err := scanRowIntoUser(rows)
		if err != nil {
			return nil, err
		}
		referrals = append(referrals, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while iterating over rows: %w", err)
	}

	return referrals, nil
}

func scanRowIntoUserForReferral(scanner *sql.Rows) (*types.User, error) {
	user := new(types.User)
	err := scanner.Scan(
		&user.ID,
		&user.Email,
		&user.ReferrerID,
		&user.ReferralCode,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}
