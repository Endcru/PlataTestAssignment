package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"time"

	"github.com/Endcru/PlataTestAssignment/internal/config"
)

// Canstant for the statements names
const (
	createQuotation = "createQuotation"
	createQuotationUpdate = "createQuotationUpdate"
	createQuotationRequest = "createQuotationRequest"
	getQuotation = "getQuotation"
	getQuotationUpdate = "getQuotationUpdate"
	getQuotationRequest = "getQuotationRequest"
	updateQuotation = "updateQuotation"
	doneQuotationRequest = "doneQuotationRequest"
)

type PostgresStorage struct {
	cfg *config.Config
    db *sql.DB
	statements map[string]*sql.Stmt
}



func (storage *PostgresStorage) AddStatement(name string, query string) error {
	stmt, err := storage.db.Prepare(query)
	if err != nil {
		return err
	}
	storage.statements[name] = stmt
	return nil
}

func AddBaseStatements(storage *PostgresStorage) error {
	const op = "storage.AddBaseStatements"

	statements := map[string]string{
		createQuotation: `
			INSERT INTO quotation (name, updated_at, rate)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO UPDATE SET updated_at = $2, rate = $3
		`,
		createQuotationUpdate: `
			INSERT INTO quotation_update (name, updated_at, previous_rate, new_rate, source)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name, updated_at) DO UPDATE SET previous_rate = $3, new_rate = $4, source = $5
		`,
		createQuotationRequest: `
			INSERT INTO quotation_request (quotation_name, requested_at)
			VALUES ($1, $2)
		`,
		getQuotation: `
			SELECT updated_at, rate FROM quotation WHERE name = $1
		`,
		getQuotationUpdate: `
			SELECT updated_at, previous_rate, new_rate, source FROM quotation_update WHERE name = $1
		`,
		getQuotationRequest: `
			SELECT quotation_name, requested_at, completed_at, done FROM quotation_request WHERE id = $1
		`,
		updateQuotation: `
			UPDATE quotation SET rate = $1, updated_at = $2 WHERE name = $3
		`,
		doneQuotationRequest: `
			UPDATE quotation_request SET done = TRUE, completed_at = $1 WHERE id = $2
		`,
	}
	for name, query := range statements {
		err := storage.AddStatement(name, query)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	return nil
}

func NewPostgresStorage(cfg *config.Config) (*PostgresStorage, error) {
	const op = "storage.NewPostgresStorage"

	db, err := sql.Open("postgres", cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS quotation (
			name VARCHAR(7) UNIQUE NOT NULL PRIMARY KEY,
			updated_at TIMESTAMP NOT NULL,
			rate FLOAT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS quotation_update (
			id SERIAL PRIMARY KEY,
			name VARCHAR(7) NOT NULL REFERENCES quotation (name),
			updated_at TIMESTAMP NOT NULL,
			previous_rate FLOAT NOT NULL,
			new_rate FLOAT NOT NULL,
			source VARCHAR(255) NOT NULL,
			FOREIGN KEY (name) REFERENCES quotation (name)
		);

		CREATE TABLE IF NOT EXISTS quotation_request (
			id SERIAL PRIMARY KEY,
			quotation_name VARCHAR(7) NOT NULL REFERENCES quotation (name),
			done BOOLEAN NOT NULL DEFAULT FALSE,
			requested_at TIMESTAMP NOT NULL,
			completed_at TIMESTAMP NULL
		);
	`)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PostgresStorage{cfg: cfg, db: db, statements: make(map[string]*sql.Stmt)}, nil
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}

func (s *PostgresStorage) CreateQuotation(name string, rate float64,) error {
	const op = "storage.CreateQuotation"

	_, err := s.statements[createQuotation].Exec(name, time.Now(), rate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *PostgresStorage) CreateQuotationUpdate(name string, previousRate float64, newRate float64, source string) error {
	const op = "storage.CreateQuotationUpdate"

	_, err := s.statements[createQuotationUpdate].Exec(name, time.Now(), previousRate, newRate, source)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *PostgresStorage) CreateQuotationRequest(name string) error {
	const op = "storage.CreateQuotationRequest"

	_, err := s.statements[createQuotationRequest].Exec(name, time.Now())
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *PostgresStorage) GetQuotation( name string) (float64, time.Time, error) {
	const op = "storage.GetQuotation"

	row := s.statements[getQuotation].QueryRow(name)
	var rate float64
	var updatedAt time.Time
	err := row.Scan(&rate, &updatedAt)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("%s: %w", op, err)
	}
	return rate, updatedAt, nil
}

func (s *PostgresStorage) GetQuotationUpdate( name string) (time.Time, float64, float64, string, error) {
	const op = "storage.GetQuotationUpdate"

	row := s.statements[getQuotationUpdate].QueryRow(name)
	var updatedAt time.Time
	var previousRate float64
	var newRate float64
	var source string
	err := row.Scan(&updatedAt, &previousRate, &newRate, &source)
	if err != nil {
		return time.Time{}, 0, 0, "", fmt.Errorf("%s: %w", op, err)
	}
	return updatedAt, previousRate, newRate, source, nil
}


func (s *PostgresStorage) GetQuotationRequest(id int) (string, time.Time, time.Time, bool, error) {
	const op = "storage.GetQuotationRequest"

	row := s.statements[getQuotationRequest].QueryRow(id)
	var quotationName string
	var requestedAt time.Time
	var completedAt time.Time
	var done bool
	err := row.Scan(&quotationName, &requestedAt, &completedAt, &done)
	if err != nil {
		return "", time.Time{}, time.Time{}, false, fmt.Errorf("%s: %w", op, err)
	}
	return quotationName, requestedAt, completedAt, done, nil
}

func (s *PostgresStorage) UpdateQuotation(name string, rate float64) error {
	const op = "storage.UpdateQuotation"

	_, err := s.statements[updateQuotation].Exec(rate, time.Now(), name)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *PostgresStorage) DoneQuotationRequest(id int) error {
	const op = "storage.DoneQuotationRequest"

	_, err := s.statements[doneQuotationRequest].Exec(time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}