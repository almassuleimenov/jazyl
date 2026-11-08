package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/almasik/jazyl/internal/domain"
)

// ---------- doctors ----------

// ListDoctors lists a clinic's doctors.
func (s *Store) ListDoctors(ctx context.Context, clinicID uuid.UUID, activeOnly bool) ([]*domain.Doctor, error) {
	rows, err := s.db.Query(ctx, `SELECT id, clinic_id, name, specialty, is_active, created_at FROM doctors
		WHERE clinic_id=$1 AND (NOT $2 OR is_active) ORDER BY name`, clinicID, activeOnly)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.Doctor, error) {
		var d domain.Doctor
		err := r.Scan(&d.ID, &d.ClinicID, &d.Name, &d.Specialty, &d.IsActive, &d.CreatedAt)
		return &d, err
	})
}

// UpsertDoctor creates (id == Nil) or updates a doctor.
func (s *Store) UpsertDoctor(ctx context.Context, d *domain.Doctor) (*domain.Doctor, error) {
	var out domain.Doctor
	var err error
	if d.ID == uuid.Nil {
		err = s.db.QueryRow(ctx, `INSERT INTO doctors (clinic_id, name, specialty, is_active) VALUES ($1,$2,$3,$4)
			RETURNING id, clinic_id, name, specialty, is_active, created_at`, d.ClinicID, d.Name, d.Specialty, d.IsActive).
			Scan(&out.ID, &out.ClinicID, &out.Name, &out.Specialty, &out.IsActive, &out.CreatedAt)
	} else {
		err = s.db.QueryRow(ctx, `UPDATE doctors SET name=$3, specialty=$4, is_active=$5 WHERE id=$2 AND clinic_id=$1
			RETURNING id, clinic_id, name, specialty, is_active, created_at`, d.ClinicID, d.ID, d.Name, d.Specialty, d.IsActive).
			Scan(&out.ID, &out.ClinicID, &out.Name, &out.Specialty, &out.IsActive, &out.CreatedAt)
	}
	return &out, mapErr(err)
}

// ---------- services ----------

// ListServices lists a clinic's price list.
func (s *Store) ListServices(ctx context.Context, clinicID uuid.UUID, activeOnly bool) ([]*domain.Service, error) {
	rows, err := s.db.Query(ctx, `SELECT id, clinic_id, name, duration_min, price_kzt, is_active, created_at FROM services
		WHERE clinic_id=$1 AND (NOT $2 OR is_active) ORDER BY name`, clinicID, activeOnly)
	if err != nil {
		return nil, mapErr(err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (*domain.Service, error) {
		var v domain.Service
		err := r.Scan(&v.ID, &v.ClinicID, &v.Name, &v.DurationMin, &v.PriceKZT, &v.IsActive, &v.CreatedAt)
		return &v, err
	})
}

// GetService returns one service of a clinic.
func (s *Store) GetService(ctx context.Context, clinicID, id uuid.UUID) (*domain.Service, error) {
	var v domain.Service
	err := s.db.QueryRow(ctx, `SELECT id, clinic_id, name, duration_min, price_kzt, is_active, created_at FROM services
		WHERE clinic_id=$1 AND id=$2`, clinicID, id).
		Scan(&v.ID, &v.ClinicID, &v.Name, &v.DurationMin, &v.PriceKZT, &v.IsActive, &v.CreatedAt)
	return &v, mapErr(err)
}

// UpsertService creates (id == Nil) or updates a service.
func (s *Store) UpsertService(ctx context.Context, v *domain.Service) (*domain.Service, error) {
	var out domain.Service
	var err error
	if v.ID == uuid.Nil {
		err = s.db.QueryRow(ctx, `INSERT INTO services (clinic_id, name, duration_min, price_kzt, is_active) VALUES ($1,$2,$3,$4,$5)
			RETURNING id, clinic_id, name, duration_min, price_kzt, is_active, created_at`, v.ClinicID, v.Name, v.DurationMin, v.PriceKZT, v.IsActive).
			Scan(&out.ID, &out.ClinicID, &out.Name, &out.DurationMin, &out.PriceKZT, &out.IsActive, &out.CreatedAt)
	} else {
		err = s.db.QueryRow(ctx, `UPDATE services SET name=$3, duration_min=$4, price_kzt=$5, is_active=$6 WHERE id=$2 AND clinic_id=$1
			RETURNING id, clinic_id, name, duration_min, price_kzt, is_active, created_at`, v.ClinicID, v.ID, v.Name, v.DurationMin, v.PriceKZT, v.IsActive).
			Scan(&out.ID, &out.ClinicID, &out.Name, &out.DurationMin, &out.PriceKZT, &out.IsActive, &out.CreatedAt)
	}
	return &out, mapErr(err)
}

// ---------- patients ----------

const patientCols = `id, clinic_id, phone, name, language, consent_at, opted_out_at, created_at, updated_at`

func scanPatient(row pgx.Row) (*domain.Patient, error) {
	var p domain.Patient
	if err := row.Scan(&p.ID, &p.ClinicID, &p.Phone, &p.Name, &p.Language, &p.ConsentAt, &p.OptedOutAt, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

// UpsertPatient finds-or-creates a patient by phone. A non-empty name overwrites an empty one.
func (s *Store) UpsertPatient(ctx context.Context, clinicID uuid.UUID, phone, name, lang string) (*domain.Patient, error) {
	if lang == "" {
		lang = "ru"
	}
	return scanPatient(s.db.QueryRow(ctx, `
		INSERT INTO patients (clinic_id, phone, name, language) VALUES ($1,$2,$3,$4)
		ON CONFLICT (clinic_id, phone) DO UPDATE
			SET name = CASE WHEN patients.name = '' THEN EXCLUDED.name ELSE patients.name END,
			    updated_at = now()
		RETURNING `+patientCols, clinicID, phone, name, lang))
}

// GetPatient returns a patient of a clinic.
func (s *Store) GetPatient(ctx context.Context, clinicID, id uuid.UUID) (*domain.Patient, error) {
	return scanPatient(s.db.QueryRow(ctx, `SELECT `+patientCols+` FROM patients WHERE clinic_id=$1 AND id=$2`, clinicID, id))
}

// GetPatientByPhone returns a patient by phone.
func (s *Store) GetPatientByPhone(ctx context.Context, clinicID uuid.UUID, phone string) (*domain.Patient, error) {
	return scanPatient(s.db.QueryRow(ctx, `SELECT `+patientCols+` FROM patients WHERE clinic_id=$1 AND phone=$2`, clinicID, phone))
}

// ListPatients lists patients, optionally filtered by name/phone substring.
func (s *Store) ListPatients(ctx context.Context, clinicID uuid.UUID, search string, page Page) ([]*domain.Patient, error) {
	ts, id, err := page.keysetArgs()
	if err != nil {
		return nil, err
	}
	rows, err := s.db.Query(ctx, `SELECT `+patientCols+` FROM patients
		WHERE clinic_id=$1 AND ($2 = '' OR phone LIKE '%'||$2||'%' OR name ILIKE '%'||$2||'%')
		  AND (created_at, id) < ($3, $4)
		ORDER BY created_at DESC, id DESC LIMIT $5`, clinicID, search, ts, id, page.limit())
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []*domain.Patient
	for rows.Next() {
		p, err := scanPatient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, mapErr(rows.Err())
}

// UpdatePatient edits name/language.
func (s *Store) UpdatePatient(ctx context.Context, clinicID, id uuid.UUID, name, lang string) (*domain.Patient, error) {
	return scanPatient(s.db.QueryRow(ctx, `UPDATE patients SET name=$3, language=$4, updated_at=now()
		WHERE clinic_id=$1 AND id=$2 RETURNING `+patientCols, clinicID, id, name, lang))
}

// SetPatientOptOut marks a patient as opted out of (or back into) messages.
func (s *Store) SetPatientOptOut(ctx context.Context, clinicID uuid.UUID, phone string, optedOut bool) error {
	var at *time.Time
	if optedOut {
		now := time.Now()
		at = &now
	}
	_, err := s.db.Exec(ctx, `UPDATE patients SET opted_out_at=$3, updated_at=now() WHERE clinic_id=$1 AND phone=$2`, clinicID, phone, at)
	return mapErr(err)
}

// SetPatientConsent records consent to receive messages.
func (s *Store) SetPatientConsent(ctx context.Context, clinicID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE patients SET consent_at=COALESCE(consent_at, now()), updated_at=now() WHERE clinic_id=$1 AND id=$2`, clinicID, id)
	return mapErr(err)
}
