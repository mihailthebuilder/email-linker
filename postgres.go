package main

import (
	"context"
	"fmt"
	"log"

	pgx "github.com/jackc/pgx/v5"
)

func getPostgresDb() *Postgres {
	var err error

	db, err := pgx.Connect(context.Background(), getEnv("DATABASE_URL"))
	if err != nil {
		log.Fatal("couldn't open database connection: ", err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}

	return &Postgres{client: db}
}

type Postgres struct {
	client *pgx.Conn
}

func (p *Postgres) Close() error {
	return p.client.Close(context.Background())
}

func (p *Postgres) UserIdExists(ctx context.Context, userId string) (bool, error) {
	var count int
	err := p.client.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE user_id = $1", userId).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("count query error: %s", err)
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (p *Postgres) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	err := p.client.QueryRow(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)

	if err != nil {
		return false, fmt.Errorf("count query error: %s", err)
	}

	if count > 0 {
		return true, nil
	}

	return false, nil
}

func (p *Postgres) CreateUser(ctx context.Context, request CreateUserRequest) error {
	_, err := p.client.Exec(ctx, "INSERT INTO users (email, password_hash, verification_code) VALUES ($1, $2, $3)", request.Email, request.Password, request.VerificationCode)
	if err != nil {
		return fmt.Errorf("error inserting in users table: %s", err)
	}

	return nil
}

func (p *Postgres) GetUser(ctx context.Context, email string) (ResultForGetUserRequest, error) {
	result := ResultForGetUserRequest{}
	err := p.client.QueryRow(ctx, "SELECT user_id, password_hash, is_verified FROM users WHERE email = $1", email).Scan(&result.User.Id, &result.User.HashedPassword, &result.User.IsVerified)

	if err == pgx.ErrNoRows {
		result.Found = false
		return result, nil
	}

	if err != nil {
		return result, fmt.Errorf("error fetching password hash: %s", err)
	}

	result.Found = true

	return result, nil
}

func (p *Postgres) AddRedirect(ctx context.Context, request AddRedirectRequest) error {
	sql := `
		INSERT INTO links (user_id, original_url, redirect_path, email_subject)
		VALUES ($1, $2, $3, $4)
	`
	tag, err := p.client.Exec(ctx, sql, request.UserId, request.Url, request.Path, request.EmailSubject)

	if err != nil {
		return err
	}

	if affected := tag.RowsAffected(); affected != 1 {
		return fmt.Errorf("expected 1 row to be affected, got %d", affected)
	}

	return nil
}

func (p *Postgres) GetRedirectRecord(ctx context.Context, path string) (RedirectRecord, error) {
	record := RedirectRecord{}
	sql := `
		SELECT l.original_url, u.email, l.number_of_times_clicked, l.email_subject 
		FROM links l
		JOIN users u on l.user_id = u.user_id
		WHERE redirect_path = $1
	`
	err := p.client.QueryRow(ctx, sql, path).Scan(&record.RedirectUrl, &record.UserEmail, &record.NumberOfTimesClicked, &record.EmailSubject)
	return record, err
}

func (p *Postgres) IncrementNumberOfTimesLinkClicked(ctx context.Context, path string) error {
	sql := `UPDATE links SET number_of_times_clicked = number_of_times_clicked + 1 WHERE redirect_path = $1`

	tag, err := p.client.Exec(ctx, sql, path)
	if err != nil {
		return err
	}

	if affected := tag.RowsAffected(); affected != 1 {
		return fmt.Errorf("expected 1 row to be affected, got %d", affected)
	}

	return nil
}

func (p *Postgres) ConfirmEmailVerified(ctx context.Context, code string) error {
	sql := `UPDATE users SET is_verified = true WHERE verification_code = $1`

	tag, err := p.client.Exec(ctx, sql, code)
	if err != nil {
		return err
	}

	if affected := tag.RowsAffected(); affected != 1 {
		return fmt.Errorf("expected 1 row to be affected, got %d", affected)
	}

	return nil
}
