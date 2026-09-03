package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"fmt"
	"time"

	"github.com/Endcru/PlataTestAssignment/internal/config"
	model "github.com/Endcru/PlataTestAssignment/internal/models"
	storage "github.com/Endcru/PlataTestAssignment/internal/storage"
)

// Canstant for the statements names
const (
	createQuotation = "createQuotation"
	createQuotationUpdate = "createQuotationUpdate"
	createQuotationRequest = "createQuotationRequest"
	getQuotation = "getQuotation"
	getQuotationUpdates = "getQuotationUpdates"
	getQuotationRequest = "getQuotationRequest"
	getQuotationRequestsUncompleted = "getQuotationRequestsUncompleted"
	updateQuotation = "updateQuotation"
	doneQuotationRequest = "doneQuotationRequest"
	deleteQuotationRequest = "deleteQuotationRequest"
)

type Storage struct {
	cfg *config.Config
    db *sql.DB
	statements map[string]*sql.Stmt
}



func (storage *Storage) AddStatement(name string, query string) error {
	stmt, err := storage.db.Prepare(query)
	if err != nil {
		return err
	}
	storage.statements[name] = stmt
	return nil
}

func AddBasePostgresStatements(storage *Storage) error {
	const op = "storage.AddBasePostgresStatements"

	statements := map[string]string{
		createQuotation: `
			INSERT INTO quotation (name, updated_at, rate)
			VALUES ($1, $2, $3)
			ON CONFLICT (name) DO NOTHING
		`,
		createQuotationUpdate: `
			INSERT INTO quotation_update (name, updated_at, previous_rate, new_rate, source)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (name, updated_at) DO UPDATE SET previous_rate = $3, new_rate = $4, source = $5
		`,
		createQuotationRequest: `
			INSERT INTO quotation_request (quotation_name, requested_at)
			VALUES ($1, $2)
			RETURNING id
		`,
		getQuotation: `
			SELECT name, updated_at, rate FROM quotation WHERE name = $1
		`,
		getQuotationUpdates: `
			SELECT id, name, updated_at, previous_rate, new_rate, source FROM quotation_update WHERE name = $1 ORDER BY updated_at DESC
		`,
		getQuotationRequest: `
			SELECT quotation_name, requested_at, COALESCE(completed_at, TIMESTAMP '0001-01-01'), done FROM quotation_request WHERE id = $1
		`,
		getQuotationRequestsUncompleted: `
			SELECT id, quotation_name, requested_at, COALESCE(completed_at, TIMESTAMP '0001-01-01'), done FROM quotation_request WHERE done = FALSE
		`,
		updateQuotation: `
			UPDATE quotation SET rate = $1, updated_at = $2 WHERE name = $3
		`,
		doneQuotationRequest: `
			UPDATE quotation_request SET done = TRUE, completed_at = $1 WHERE id = $2
		`,
		deleteQuotationRequest: `
			DELETE FROM quotation_request WHERE id = $1
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

func NewPostgresStorage(cfg *config.Config) (*Storage, error) {
	const op = "storage.NewPostgresStorage"

	db, err := sql.Open("postgres", cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.Exec(`
		DROP TABLE IF EXISTS quotation_request, quotation_update, quotation CASCADE;

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
			UNIQUE (name, updated_at)
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

	return &Storage{cfg: cfg, db: db, statements: make(map[string]*sql.Stmt)}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) CreateQuotation(quotation model.Quotation) error {
	const op = "storage.CreateQuotation"

	res, err := s.statements[createQuotation].Exec(quotation.Name, quotation.UpdatedAt, quotation.Rate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return storage.ErrQuotationAlreadyExists
	}
	return nil
}

func (s *Storage) CreateQuotationUpdate(quotationUpdate model.QuotationUpdate) error {
	const op = "storage.CreateQuotationUpdate"

	_, err := s.statements[createQuotationUpdate].Exec(quotationUpdate.Name, quotationUpdate.UpdatedAt, quotationUpdate.PreviousRate, quotationUpdate.NewRate, quotationUpdate.Source)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) CreateQuotationRequest(quotationRequest model.QuotationRequest) (int, error) {
	const op = "storage.CreateQuotationRequest"

	var id int
	err := s.statements[createQuotationRequest].QueryRow(quotationRequest.Name, quotationRequest.RequestedAt).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return id, nil
}

func (s *Storage) GetQuotation(name string) (quotation model.Quotation, err error) {
	const op = "storage.GetQuotation"

	row := s.statements[getQuotation].QueryRow(name)
	err = row.Scan(&quotation.Name, &quotation.UpdatedAt, &quotation.Rate)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.Quotation{}, storage.ErrQuotationNotFound
		}
		return model.Quotation{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotation, nil
}

func (s *Storage) GetQuotationUpdates(name string) (quotationUpdates []model.QuotationUpdate, err error) {
	const op = "storage.GetQuotationUpdates"

	rows, err := s.statements[getQuotationUpdates].Query(name)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var u model.QuotationUpdate
		err = rows.Scan(&u.ID, &u.Name, &u.UpdatedAt, &u.PreviousRate, &u.NewRate, &u.Source)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		quotationUpdates = append(quotationUpdates, u)
	}
	return quotationUpdates, nil
}


func (s *Storage) GetQuotationRequest(id int) (quotationRequest model.QuotationRequest, err error) {
	const op = "storage.GetQuotationRequest"

	row := s.statements[getQuotationRequest].QueryRow(id)
	err = row.Scan(&quotationRequest.Name, &quotationRequest.RequestedAt, &quotationRequest.CompletedAt, &quotationRequest.Done)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.QuotationRequest{}, storage.ErrQuotationRequestNotFound
		}
		return model.QuotationRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	return quotationRequest, nil
}

func (s *Storage) GetQuotationRequestsUncompleted() (quotationRequests []model.QuotationRequest, err error) {
	const op = "storage.GetQuotationRequestsUncompleted"

	rows, err := s.statements[getQuotationRequestsUncompleted].Query()
	if err != nil {
		return []model.QuotationRequest{}, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var quotationRequest model.QuotationRequest
		err = rows.Scan(&quotationRequest.ID, &quotationRequest.Name, &quotationRequest.RequestedAt, &quotationRequest.CompletedAt, &quotationRequest.Done)
		if err != nil {
			return []model.QuotationRequest{}, fmt.Errorf("%s: %w", op, err)
		}
		quotationRequests = append(quotationRequests, quotationRequest)
	}
	return quotationRequests, nil
}

func (s *Storage) UpdateQuotation(quotation model.Quotation) error {
	const op = "storage.UpdateQuotation"

	_, err := s.statements[updateQuotation].Exec(quotation.Rate, quotation.UpdatedAt, quotation.Name)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DoneQuotationRequest(id int) error {
	const op = "storage.DoneQuotationRequest"

	_, err := s.statements[doneQuotationRequest].Exec(time.Now(), id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) DeleteQuotationRequest(id int) error {
	const op = "storage.DeleteQuotationRequest"

	_, err := s.statements[deleteQuotationRequest].Exec(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}