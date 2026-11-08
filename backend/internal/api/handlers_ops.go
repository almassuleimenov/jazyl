package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/almasik/jazyl/internal/auth"
	"github.com/almasik/jazyl/internal/core"
	"github.com/almasik/jazyl/internal/domain"
	"github.com/almasik/jazyl/internal/phone"
	"github.com/almasik/jazyl/internal/store"
)

func clinicParam(r *http.Request) uuid.UUID {
	id, _ := uuid.Parse(chi.URLParam(r, "clinicID")) // validated by requireClinicAccess
	return id
}

func optionalID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	if chi.URLParam(r, "id") == "" {
		return uuid.Nil, true
	}
	return pathUUID(w, r, "id")
}

// ---------- doctors & services ----------

func (a *API) listDoctors(w http.ResponseWriter, r *http.Request) {
	ds, err := a.store.ListDoctors(r.Context(), clinicParam(r), r.URL.Query().Get("active") == "true")
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(ds)})
}

func (a *API) upsertDoctor(w http.ResponseWriter, r *http.Request) {
	id, ok := optionalID(w, r)
	if !ok {
		return
	}
	var in struct {
		Name      string `json:"name"`
		Specialty string `json:"specialty"`
		IsActive  *bool  `json:"is_active"`
	}
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "name is required")
		return
	}
	d, err := a.store.UpsertDoctor(r.Context(), &domain.Doctor{
		ID: id, ClinicID: clinicParam(r), Name: strings.TrimSpace(in.Name), Specialty: strings.TrimSpace(in.Specialty), IsActive: in.IsActive == nil || *in.IsActive,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, statusFor(id), d)
}

func (a *API) listServices(w http.ResponseWriter, r *http.Request) {
	ss, err := a.store.ListServices(r.Context(), clinicParam(r), r.URL.Query().Get("active") == "true")
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(ss)})
}

func (a *API) upsertService(w http.ResponseWriter, r *http.Request) {
	id, ok := optionalID(w, r)
	if !ok {
		return
	}
	var in struct {
		Name        string `json:"name"`
		DurationMin int    `json:"duration_min"`
		PriceKZT    int    `json:"price_kzt"`
		IsActive    *bool  `json:"is_active"`
	}
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Name) == "" {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "name is required")
		return
	}
	if in.DurationMin == 0 {
		in.DurationMin = 30
	}
	v, err := a.store.UpsertService(r.Context(), &domain.Service{
		ID: id, ClinicID: clinicParam(r), Name: strings.TrimSpace(in.Name), DurationMin: in.DurationMin, PriceKZT: in.PriceKZT,
		IsActive: in.IsActive == nil || *in.IsActive,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, statusFor(id), v)
}

func statusFor(id uuid.UUID) int {
	if id == uuid.Nil {
		return http.StatusCreated
	}
	return http.StatusOK
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}

// ---------- patients ----------

func (a *API) listPatients(w http.ResponseWriter, r *http.Request) {
	page := store.Page{Limit: queryInt(r, "limit", 50), Cursor: r.URL.Query().Get("cursor")}
	ps, err := a.store.ListPatients(r.Context(), clinicParam(r), strings.TrimSpace(r.URL.Query().Get("q")), page)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(ps, page.Limit, func(p *domain.Patient) string { return store.MakeCursor(p.CreatedAt, p.ID) }))
}

func (a *API) getPatient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	cid := clinicParam(r)
	p, err := a.store.GetPatient(r.Context(), cid, id)
	if err != nil {
		fail(w, r, err)
		return
	}
	appts, err := a.store.ListAppointments(r.Context(), cid, store.ApptFilter{
		From: time.Now().AddDate(-2, 0, 0), To: time.Now().AddDate(1, 0, 0), PatientID: &p.ID,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"patient": p, "appointments": nonNil(appts)})
}

func (a *API) updatePatient(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Name     *string `json:"name"`
		Language *string `json:"language"`
		OptedOut *bool   `json:"opted_out"`
	}
	if !decode(w, r, &in) {
		return
	}
	cid := clinicParam(r)
	p, err := a.store.GetPatient(r.Context(), cid, id)
	if err != nil {
		fail(w, r, err)
		return
	}
	name, lang := p.Name, p.Language
	if in.Name != nil {
		name = strings.TrimSpace(*in.Name)
	}
	if in.Language != nil {
		if *in.Language != "ru" && *in.Language != "kk" {
			writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "language must be ru or kk")
			return
		}
		lang = *in.Language
	}
	if in.OptedOut != nil {
		if err := a.store.SetPatientOptOut(r.Context(), cid, p.Phone, *in.OptedOut); err != nil {
			fail(w, r, err)
			return
		}
	}
	out, err := a.store.UpdatePatient(r.Context(), cid, id, name, lang)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// ---------- appointments ----------

func (a *API) clinicLoc(r *http.Request) *time.Location {
	if c, err := a.store.GetClinic(r.Context(), clinicParam(r)); err == nil {
		return c.Location()
	}
	return time.FixedZone("UTC+5", 5*3600)
}

func (a *API) listAppointments(w http.ResponseWriter, r *http.Request) {
	loc := a.clinicLoc(r)
	from, to, err := queryRange(r, loc, 1)
	if err != nil {
		fail(w, r, err)
		return
	}
	if r.URL.Query().Get("from") == "" && r.URL.Query().Get("to") == "" {
		// Default: today and tomorrow.
		n := time.Now().In(loc)
		from = time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
		to = from.AddDate(0, 0, 2)
	}
	doctorID, err := queryUUID(r, "doctor_id")
	if err != nil {
		fail(w, r, err)
		return
	}
	patientID, err := queryUUID(r, "patient_id")
	if err != nil {
		fail(w, r, err)
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !domain.ValidApptStatus(status) {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "status")
		return
	}
	as, err := a.store.ListAppointments(r.Context(), clinicParam(r), store.ApptFilter{From: from, To: to, Status: status, DoctorID: doctorID, PatientID: patientID})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(as), "from": from, "to": to})
}

type appointmentInput struct {
	PatientPhone string     `json:"patient_phone"`
	PatientName  string     `json:"patient_name"`
	PatientLang  string     `json:"patient_language"`
	DoctorID     *uuid.UUID `json:"doctor_id"`
	ServiceID    *uuid.UUID `json:"service_id"`
	ServiceName  string     `json:"service_name"`
	StartsAt     time.Time  `json:"starts_at"`
	DurationMin  int        `json:"duration_min"`
	Notes        string     `json:"notes"`
}

func (a *API) createAppointment(w http.ResponseWriter, r *http.Request) {
	var in appointmentInput
	if !decode(w, r, &in) {
		return
	}
	if in.StartsAt.Before(time.Now().Add(-24 * time.Hour)) {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "starts_at is too far in the past")
		return
	}
	if in.DurationMin < 0 || in.DurationMin > 480 {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "duration_min must be 5..480")
		return
	}
	cid := clinicParam(r)
	appt, created, err := a.svc.CreateAppointment(r.Context(), cid, core.NewAppointment{
		PatientPhone: in.PatientPhone, PatientName: in.PatientName, PatientLang: in.PatientLang, DoctorID: in.DoctorID,
		ServiceID: in.ServiceID, ServiceName: in.ServiceName, StartsAt: in.StartsAt, DurationMin: in.DurationMin,
		Notes: in.Notes, Source: "manual",
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	writeJSON(w, status, appt)
}

func (a *API) getAppointment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	appt, err := a.store.GetAppointment(r.Context(), clinicParam(r), id)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, appt)
}

func (a *API) updateAppointment(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		StartsAt    *time.Time `json:"starts_at"`
		DurationMin *int       `json:"duration_min"`
		DoctorID    *uuid.UUID `json:"doctor_id"`
		ClearDoctor bool       `json:"clear_doctor"`
		ServiceID   *uuid.UUID `json:"service_id"`
		ServiceName *string    `json:"service_name"`
		Notes       *string    `json:"notes"`
	}
	if !decode(w, r, &in) {
		return
	}
	if in.DurationMin != nil && (*in.DurationMin < 5 || *in.DurationMin > 480) {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "duration_min must be 5..480")
		return
	}
	appt, err := a.svc.UpdateAppointment(r.Context(), clinicParam(r), id, core.AppointmentPatch{
		StartsAt: in.StartsAt, DurationMin: in.DurationMin, DoctorID: in.DoctorID, ClearDoctor: in.ClearDoctor,
		ServiceID: in.ServiceID, ServiceName: in.ServiceName, Notes: in.Notes,
	})
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, appt)
}

func (a *API) setAppointmentStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &in) {
		return
	}
	appt, err := a.svc.ChangeStatus(r.Context(), clinicParam(r), id, in.Status, "staff")
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, appt)
}

func (a *API) importSchedule(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Text string `json:"text"`
	}
	if !decode(w, r, &in) {
		return
	}
	if strings.TrimSpace(in.Text) == "" || len(in.Text) > 20000 {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "text is required (max 20000 chars)")
		return
	}
	ack, err := a.svc.ImportSchedule(r.Context(), clinicParam(r), "dashboard", in.Text)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"summary": ack})
}

// ---------- leads ----------

func (a *API) listLeads(w http.ResponseWriter, r *http.Request) {
	page := store.Page{Limit: queryInt(r, "limit", 50), Cursor: r.URL.Query().Get("cursor")}
	ls, err := a.store.ListLeads(r.Context(), clinicParam(r), r.URL.Query().Get("status"), page)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(ls, page.Limit, func(l *domain.Lead) string { return store.MakeCursor(l.CreatedAt, l.ID) }))
}

func (a *API) updateLead(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		Status string `json:"status"`
	}
	if !decode(w, r, &in) {
		return
	}
	switch in.Status {
	case "new", "contacted", "rejected":
	default:
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "status must be new, contacted or rejected (use /convert to book)")
		return
	}
	if err := a.store.UpdateLeadStatus(r.Context(), clinicParam(r), id, in.Status, nil); err != nil {
		fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) convertLead(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	var in struct {
		StartsAt  time.Time  `json:"starts_at"`
		DoctorID  *uuid.UUID `json:"doctor_id"`
		ServiceID *uuid.UUID `json:"service_id"`
	}
	if !decode(w, r, &in) {
		return
	}
	appt, err := a.svc.ConvertLead(r.Context(), clinicParam(r), id, in.StartsAt, in.DoctorID, in.ServiceID)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, appt)
}

// ---------- messages, chats, reviews ----------

func (a *API) listMessages(w http.ResponseWriter, r *http.Request) {
	ph := ""
	if raw := r.URL.Query().Get("phone"); raw != "" {
		p, err := phone.Normalize(raw)
		if err != nil {
			writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "phone")
			return
		}
		ph = p
	}
	page := store.Page{Limit: queryInt(r, "limit", 50), Cursor: r.URL.Query().Get("cursor")}
	ms, err := a.store.ListMessages(r.Context(), clinicParam(r), ph, page)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(ms, page.Limit, func(m *domain.Message) string { return store.MakeCursor(m.CreatedAt, m.ID) }))
}

// sendManualMessage lets staff write to a patient from the dashboard (goes through the outbox pacing).
func (a *API) sendManualMessage(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Phone string `json:"phone"`
		Text  string `json:"text"`
	}
	if !decode(w, r, &in) {
		return
	}
	ph, err := phone.Normalize(in.Phone)
	if err != nil {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "phone")
		return
	}
	text := strings.TrimSpace(in.Text)
	if text == "" || len([]rune(text)) > 4000 {
		writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "text is required (max 4000 chars)")
		return
	}
	cid := clinicParam(r)
	var pid *uuid.UUID
	if p, err := a.store.GetPatientByPhone(r.Context(), cid, ph); err == nil {
		if p.OptedOutAt != nil {
			writeErr(w, r, http.StatusConflict, "opted_out", "patient unsubscribed from messages")
			return
		}
		pid = &p.ID
	} else if !errors.Is(err, domain.ErrNotFound) {
		fail(w, r, err)
		return
	}
	m, err := a.store.EnqueueMessage(r.Context(), store.OutMessage{ClinicID: cid, PatientID: pid, Phone: ph, Kind: domain.MsgManual, Body: text})
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "message.manual", "message", m.ID.String(), nil)
	writeJSON(w, http.StatusAccepted, m)
}

func (a *API) listChats(w http.ResponseWriter, r *http.Request) {
	from, to, err := queryRange(r, a.clinicLoc(r), 30)
	if err != nil {
		fail(w, r, err)
		return
	}
	cs, err := a.store.ListChats(r.Context(), clinicParam(r), from, to, queryInt(r, "limit", 100))
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(cs), "from": from, "to": to})
}

func (a *API) listReviews(w http.ResponseWriter, r *http.Request) {
	page := store.Page{Limit: queryInt(r, "limit", 50), Cursor: r.URL.Query().Get("cursor")}
	rs, err := a.store.ListReviews(r.Context(), clinicParam(r), page)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, newPage(rs, page.Limit, func(v *domain.Review) string { return store.MakeCursor(v.CreatedAt, v.ID) }))
}

// ---------- analytics & reports ----------

func (a *API) analyticsSummary(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.GetClinic(r.Context(), clinicParam(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	from, to, err := queryRange(r, c.Location(), 30)
	if err != nil {
		fail(w, r, err)
		return
	}
	m, err := a.store.ClinicMetrics(r.Context(), c.ID, from, to, c.AvgCheckKZT)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}

func (a *API) analyticsDaily(w http.ResponseWriter, r *http.Request) {
	c, err := a.store.GetClinic(r.Context(), clinicParam(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	from, to, err := queryRange(r, c.Location(), 30)
	if err != nil {
		fail(w, r, err)
		return
	}
	pts, err := a.store.DailySeries(r.Context(), c.ID, from, to, c.Timezone)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(pts), "timezone": c.Timezone})
}

func (a *API) listReports(w http.ResponseWriter, r *http.Request) {
	rs, err := a.store.ListReports(r.Context(), clinicParam(r), queryInt(r, "limit", 20))
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(rs)})
}

func periodParam(r *http.Request) (string, error) {
	p := r.URL.Query().Get("period")
	if p == "" {
		p = "week"
	}
	if p != "week" && p != "month" {
		return "", fmt.Errorf("%w: period must be week or month", domain.ErrInvalid)
	}
	return p, nil
}

func (a *API) previewReport(w http.ResponseWriter, r *http.Request) {
	period, err := periodParam(r)
	if err != nil {
		fail(w, r, err)
		return
	}
	c, err := a.store.GetClinic(r.Context(), clinicParam(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	start, end := core.ReportPeriod(period, time.Now(), c.Location())
	m, err := a.store.ClinicMetrics(r.Context(), c.ID, start, end, c.AvgCheckKZT)
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"text": core.RenderReport(c, period, m), "metrics": m})
}

func (a *API) sendReport(w http.ResponseWriter, r *http.Request) {
	period, err := periodParam(r)
	if err != nil {
		fail(w, r, err)
		return
	}
	cid := clinicParam(r)
	sent, err := a.svc.SendReport(r.Context(), cid, period, true)
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "report.sent", "clinic", cid.String(), map[string]any{"period": period})
	writeJSON(w, http.StatusAccepted, map[string]bool{"queued": sent})
}

// ---------- site keys ----------

func (a *API) listSiteKeys(w http.ResponseWriter, r *http.Request) {
	ks, err := a.store.ListSiteKeys(r.Context(), clinicParam(r))
	if err != nil {
		fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": nonNil(ks)})
}

func (a *API) createSiteKey(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name           string   `json:"name"`
		AllowedOrigins []string `json:"allowed_origins"`
	}
	if !decode(w, r, &in) {
		return
	}
	origins := make([]string, 0, len(in.AllowedOrigins))
	for _, o := range in.AllowedOrigins {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if !strings.HasPrefix(o, "https://") && !strings.HasPrefix(o, "http://localhost") {
			writeErr(w, r, http.StatusUnprocessableEntity, "invalid", "allowed_origins must be https:// origins (or http://localhost for dev)")
			return
		}
		origins = append(origins, o)
	}
	token, hash, err := auth.RandomToken("jzs_", 24)
	if err != nil {
		fail(w, r, err)
		return
	}
	cid := clinicParam(r)
	k, err := a.store.CreateSiteKey(r.Context(), cid, strings.TrimSpace(in.Name), token[:10], hash, origins)
	if err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "site_key.created", "site_key", k.ID.String(), nil)
	writeJSON(w, http.StatusCreated, map[string]any{"key": token, "site_key": k, "note": "store this key now; it is shown only once"})
}

func (a *API) revokeSiteKey(w http.ResponseWriter, r *http.Request) {
	id, ok := pathUUID(w, r, "id")
	if !ok {
		return
	}
	cid := clinicParam(r)
	if err := a.store.RevokeSiteKey(r.Context(), cid, id); err != nil {
		fail(w, r, err)
		return
	}
	a.audit(r, &cid, "site_key.revoked", "site_key", id.String(), nil)
	w.WriteHeader(http.StatusNoContent)
}
