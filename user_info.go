package data

import (
	"database/sql"
	"errors"
	"fmt"
	"gopkg.in/gomail.v2"
	"log"
	"math/rand"
	"os"
	"time"
)

var AnonymousUser = &UserInfo{}

type UserInfo struct {
	ID           int64     `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Fname        string    `json:"fname"`
	Sname        string    `json:"sname"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	UserRole     string    `json:"user_role"`
	Activated    bool      `json:"activated"`
	Version      int       `json:"version"`
	Token        Token     `json:"token"`
}

type Token struct {
	Hash   string    `json:"hash"`
	UserID int64     `json:"user_id"`
	Expiry time.Time `json:"expiry"`
}

type UserInfoModel struct {
	db *sql.DB
}

func (u *UserInfo) IsAnonymous() bool {
	return u == AnonymousUser
}

func (m *UserInfoModel) CreateUser(userInfo *UserInfo) (*UserInfo, error) {
	query := `
		INSERT INTO users (fname, sname, email, password_hash, user_role, activated, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at, fname, sname, email, user_role, activated, version`

	err := m.db.QueryRow(query,
		userInfo.Fname, userInfo.Sname, userInfo.Email, userInfo.PasswordHash,
		userInfo.UserRole, userInfo.Activated, userInfo.Version,
	).Scan(
		&userInfo.ID, &userInfo.CreatedAt, &userInfo.UpdatedAt,
		&userInfo.Fname, &userInfo.Sname, &userInfo.Email,
		&userInfo.UserRole, &userInfo.Activated, &userInfo.Version,
	)
	if err != nil {
		return nil, err
	}
	go m.sendActivationMessage(userInfo)

	return userInfo, nil
}

func (m *UserInfoModel) Update(userInfo *UserInfo) (*UserInfo, error) {
	query := `
			UPDATE users 
			SET fname=$1, sname=$2, email=$3, password_hash=$4, user_role=$5, activated=$6, version=$7
			WHERE id=$8
			RETURNING id, created_at, updated_at, fname, sname, email, user_role, activated, version`

	var newUserInfo UserInfo
	err := m.db.QueryRow(query,
		userInfo.Fname, userInfo.Sname, userInfo.Email, userInfo.PasswordHash,
		userInfo.UserRole, userInfo.Activated, userInfo.Version, userInfo.ID,
	).Scan(
		&newUserInfo.ID, &newUserInfo.CreatedAt, &newUserInfo.UpdatedAt,
		&newUserInfo.Fname, &newUserInfo.Sname, &newUserInfo.Email,
		&newUserInfo.UserRole, &newUserInfo.Activated, &newUserInfo.Version,
	)

	return &newUserInfo, err
}

func (m *UserInfoModel) GetByID(userID int64) (*UserInfo, error) {
	query := `
        SELECT u.id, u.created_at, u.updated_at, u.fname, u.sname, u.email, u.password_hash, u.user_role, u.activated, u.version, t.hash, t.expiry
        FROM users u
        INNER JOIN tokens t ON u.id = t.user_id
        WHERE u.id = $1`

	var userInfo UserInfo
	err := m.db.QueryRow(query, userID).Scan(
		&userInfo.ID, &userInfo.CreatedAt, &userInfo.UpdatedAt,
		&userInfo.Fname, &userInfo.Sname, &userInfo.Email, &userInfo.PasswordHash,
		&userInfo.UserRole, &userInfo.Activated, &userInfo.Version,
		&userInfo.Token.Hash, &userInfo.Token.Expiry,
	)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrUserInfoNotFound
		}
		return nil, err
	}

	return &userInfo, nil
}

func (m *UserInfoModel) GetByEmail(email string) (*UserInfo, error) {
	query := `
        SELECT u.id, u.created_at, u.updated_at, u.fname, u.sname, u.email, u.user_role, u.activated, u.version, t.hash, t.expiry
        FROM users u
        INNER JOIN tokens t ON u.id = t.user_id
        WHERE u.email = $1`

	var userInfo UserInfo
	err := m.db.QueryRow(query, email).Scan(
		&userInfo.ID, &userInfo.CreatedAt, &userInfo.UpdatedAt,
		&userInfo.Fname, &userInfo.Sname, &userInfo.Email,
		&userInfo.UserRole, &userInfo.Activated, &userInfo.Version,
		&userInfo.Token.Hash, &userInfo.Token.Expiry,
	)
	if err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrUserInfoNotFound
		}
		return nil, err
	}

	return &userInfo, nil
}

func (m *UserInfoModel) getAllUsers() ([]*UserInfo, error) {
	query := `
		SELECT id, created_at, updated_at, fname, sname, email, user_role, activated, version
		FROM users`

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*UserInfo
	for rows.Next() {
		var user UserInfo
		err := rows.Scan(
			&user.ID, &user.CreatedAt, &user.UpdatedAt,
			&user.Fname, &user.Sname, &user.Email,
			&user.UserRole, &user.Activated, &user.Version,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		if errors.Is(sql.ErrNoRows, err) {
			return nil, ErrUserInfoNotFound
		}
		return nil, err
	}

	return users, nil
}

func (m *UserInfoModel) updateByID(userID int64, newUserInfo *UserInfo) (*UserInfo, error) {
	userInfo, err := m.GetByID(userID)
	if err != nil {
		return nil, err
	}

	userInfo.Fname = newUserInfo.Fname
	userInfo.Sname = newUserInfo.Sname
	userInfo.Email = newUserInfo.Email
	userInfo.UserRole = newUserInfo.UserRole
	userInfo.Version = newUserInfo.Version

	return nil, nil
}

func (m *UserInfoModel) deleteByID(userID int64) error {
	query := "DELETE FROM users WHERE id = $1"
	_, err := m.db.Exec(query, userID)
	return err
}

var (
	activationCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func generateActivationCode(length int, charset string) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}

type Mail struct {
	Sender  string
	To      string
	Subject string
	Body    string
}

func (m *UserInfoModel) sendActivationMessage(userInfo *UserInfo) {
	senderEmailUsername := os.Getenv("EMAIL_USERNAME")
	senderEmailPassword := os.Getenv("EMAIL_PASSWORD")
	autoGeneratedActivationCode := generateActivationCode(32, activationCodeCharset)
	host := "smtp.mail.ru"
	port := 465

	// here
	activationLink := fmt.Sprintf("http://localhost:4000/activation?id=%d&activationCode=%s", userInfo.ID, autoGeneratedActivationCode)
	body := fmt.Sprintf(`Dear %s, your activation link is <a href="%s">%s</a>.`, userInfo.Fname, activationLink, activationLink)

	request := Mail{
		Sender:  senderEmailUsername,
		To:      userInfo.Email,
		Subject: "Account Activation Code",
		Body:    body,
	}

	msg := gomail.NewMessage()
	msg.SetHeader("From", senderEmailUsername)
	msg.SetHeader("To", userInfo.Email)
	msg.SetHeader("Subject", "Account Activation Code")
	msg.SetBody("text/html", request.Body)

	n := gomail.NewDialer(host, port, senderEmailUsername, senderEmailPassword)
	if err := n.DialAndSend(msg); err != nil {
		panic(err)
	}
	fmt.Println("Email sent successfully")

	query := "INSERT INTO tokens(hash, user_id, expiry) VALUES ($1, $2, $3)"
	expiryTime := time.Now().Add(120 * time.Second)
	_, err2 := m.db.Exec(query, autoGeneratedActivationCode, userInfo.ID, expiryTime)
	if err2 != nil {
		log.Fatal(err2)
	}
}
