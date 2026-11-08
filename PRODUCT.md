# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Backend: Go (`backend/`), PostgreSQL. Dashboard: Next.js App Router + TypeScript + Tailwind (`frontend/`), chosen by the user. Clinic websites are separate projects on Vercel and are not part of the dashboard.

## Users

- **Founder (Jazyl team, superadmin)** — on a laptop. Onboards clinics, links each clinic's WhatsApp number by QR, sets plans, creates logins for clinic staff, watches every clinic's activity and MRR. A future assistant gets the same platform role.
- **Clinic owner** — mostly on a phone, between patients. Checks the numbers (confirmations, prevented no-shows, saved revenue), reads reports and complaints, manages the team.
- **Clinic administrator** — at the reception desk computer. Works the schedule for today and tomorrow, sees who confirmed and who asked to reschedule, handles website leads.

## Product Purpose

Jazyl is a subscription service for private dental clinics in Kazakhstan (Turkestan region first) that makes every booking reach the chair: WhatsApp reminders with "1 — приду / 2 — перенести" replies, a 2GIS review flow, website bookings, and owner reports in numbers. The dashboard is the operating console for all of that. Success: an administrator sees at a glance who is coming tomorrow, and an owner sees in tenge what the service saved.

## Positioning

Turnkey service rather than a CRM to learn: Jazyl sets everything up and is accountable for results in numbers. It works on top of existing CRMs (MacDent) or a paper notebook. The bot is a strict reminder machine, not an AI chatbot.

## Operating Context

- Administrators send tomorrow's list to the bot in WhatsApp ("тетрадь" mode); the dashboard shows the result.
- Reports go to the owner in WhatsApp on Fridays and on the 1st of the month.
- Languages: Russian interface; patient messages in Russian or Kazakh.

## Capabilities and Constraints

- Roles: superadmin (all clinics), owner (own clinic, manages team), admin (own clinic, day-to-day).
- Access is granted by the founder creating a login and password for clinic staff (user's choice). Users created this way must change the password at first login.
- 2FA (TOTP, Google Authenticator) is mandatory for the founder/superadmin and optional for clinic users.
- Tariffs: «Без неявок» 15 000 ₸, «Полное кресло» 29 000 ₸, «Рост» 49 000 ₸.
- Patient data is personal and medical-adjacent: no diagnoses in messages; data residency in Kazakhstan is an open production requirement.
- Website analytics are out of scope (the founder uses Vercel analytics).

## Brand Commitments

- Name **Jazyl** (from Kazakh «жазылу» — to sign up). Logo: lowercase "jazyl", the dot of the "j" is a checkmark; never teeth, smiles, syringes or crosses.
- Colors (binding): deep teal `#0F5E63` (primary, dome of the Khoja Ahmed Yasawi mausoleum), warm amber `#F2A541` (actions), milk `#F7F5F0` (background), graphite `#1F2328` (text). White text on amber is not allowed.
- Fonts (binding): Manrope for headings, Inter for text (both support Kazakh letters).
- Voice: concrete ("3 неявки в неделю = 45 000 ₸"), local and polite (на «вы»), honest — no "№1", no invented reviews.
- Mascot «Кезекші» (a teal message bubble with an amber checkmark) lives on the clinic side: reports and staff materials.

## Evidence on Hand

No real clinic data or case studies yet (pilot starts 2026-12-01). Do not fabricate customers, testimonials or numbers in the UI.

## Product Principles

1. Numbers over adjectives: every screen answers "what happened and what is it worth".
2. Nothing to learn: an administrator understands the screen without training.
3. Tomorrow first: the default view is the schedule that still can be saved.
4. Patient data is guarded: least privilege, no data the role does not need.
5. Honest by default: show empty states and failures plainly, never fake activity.

## Accessibility & Inclusion

Owners use phones in a hurry; administrators may be non-technical. Large touch targets on mobile, readable contrast (WCAG AA), plain Russian labels.

### Dashboard style preference

For the dashboard the founder chose the familiar category standard (2026-11-08): a classic SaaS console at the craft level of Linear and Stripe, in Jazyl's brand colors and fonts, with no special visual gimmick.
