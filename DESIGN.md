---
name: Jazyl
description: The operating console for Jazyl clinics, a calm SaaS dashboard in teal, amber and milk.
colors:
  brand: "#0f5e63"
  brand-strong: "#0b4a4e"
  brand-soft: "#e3efef"
  brand-line: "#b9d5d6"
  brand-ink: "#d4e6e6"
  action: "#f2a541"
  action-hover: "#e8962b"
  action-soft: "#fcf0dc"
  action-ink: "#7a4a07"
  milk: "#f7f5f0"
  surface: "#ffffff"
  sunk: "#f1eee7"
  line: "#e6e1d6"
  line-strong: "#d3cdbf"
  ink: "#1f2328"
  ink-2: "#545960"
  ink-3: "#6b7078"
  ok: "#1d6f46"
  ok-soft: "#e4f2ea"
  bad: "#b42318"
  bad-soft: "#fbe9e7"
typography:
  display:
    fontFamily: "Manrope, Inter, ui-sans-serif, sans-serif"
    fontSize: "26px"
    fontWeight: 700
    lineHeight: 1.25
    letterSpacing: "-0.015em"
  headline:
    fontFamily: "Manrope, Inter, ui-sans-serif, sans-serif"
    fontSize: "24px"
    fontWeight: 700
    lineHeight: 1.33
    letterSpacing: "-0.015em"
  title:
    fontFamily: "Manrope, Inter, ui-sans-serif, sans-serif"
    fontSize: "15px"
    fontWeight: 600
    lineHeight: 1.4
    letterSpacing: "-0.015em"
  body:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.43
    fontFeature: "\"cv11\", \"ss01\""
  body-small:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "13px"
    fontWeight: 400
    lineHeight: 1.4
    fontFeature: "\"cv11\", \"ss01\""
  label:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 500
    lineHeight: 1.33
    fontFeature: "\"cv11\", \"ss01\""
  figure:
    fontFamily: "Manrope, Inter, ui-sans-serif, sans-serif"
    fontSize: "26px"
    fontWeight: 700
    lineHeight: 1.25
    letterSpacing: "-0.02em"
    fontFeature: "\"tnum\""
rounded:
  bar: "6px"
  control: "8px"
  panel: "12px"
  pill: "9999px"
spacing:
  xs: "6px"
  sm: "12px"
  md: "16px"
  lg: "20px"
  xl: "24px"
  2xl: "32px"
components:
  button-primary:
    backgroundColor: "{colors.action}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "40px"
  button-primary-hover:
    backgroundColor: "{colors.action-hover}"
    textColor: "{colors.ink}"
  button-brand:
    backgroundColor: "{colors.brand}"
    textColor: "{colors.surface}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "40px"
  button-brand-hover:
    backgroundColor: "{colors.brand-strong}"
  button-secondary:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "40px"
  button-secondary-hover:
    backgroundColor: "{colors.sunk}"
  button-secondary-sm:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    rounded: "{rounded.control}"
    padding: "0 12px"
    height: "32px"
  button-ghost:
    textColor: "{colors.ink-2}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "40px"
  button-danger:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.bad}"
    rounded: "{rounded.control}"
    padding: "0 16px"
    height: "40px"
  button-danger-hover:
    backgroundColor: "{colors.bad-soft}"
  input:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "{typography.body}"
    rounded: "{rounded.control}"
    padding: "0 12px"
    height: "40px"
  panel:
    backgroundColor: "{colors.surface}"
    rounded: "{rounded.panel}"
    padding: "20px"
  chip-neutral:
    backgroundColor: "{colors.sunk}"
    textColor: "{colors.ink-2}"
    typography: "{typography.label}"
    rounded: "{rounded.pill}"
    padding: "0 10px"
    height: "24px"
  chip-teal:
    backgroundColor: "{colors.brand-soft}"
    textColor: "{colors.brand}"
    rounded: "{rounded.pill}"
    height: "24px"
  chip-amber:
    backgroundColor: "{colors.action-soft}"
    textColor: "{colors.action-ink}"
    rounded: "{rounded.pill}"
    height: "24px"
  chip-red:
    backgroundColor: "{colors.bad-soft}"
    textColor: "{colors.bad}"
    rounded: "{rounded.pill}"
    height: "24px"
  chip-green:
    backgroundColor: "{colors.ok-soft}"
    textColor: "{colors.ok}"
    rounded: "{rounded.pill}"
    height: "24px"
  sidebar:
    backgroundColor: "{colors.brand}"
    textColor: "{colors.brand-ink}"
    width: "240px"
  sidebar-item-active:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.brand}"
    rounded: "{rounded.control}"
    height: "40px"
  phone-topbar:
    backgroundColor: "{colors.brand}"
    textColor: "{colors.brand-ink}"
    height: "56px"
  save-bar:
    backgroundColor: "{colors.milk}"
    padding: "12px 0"
---

# Design System: Jazyl

## Overview

**Creative North Star: "The Reception Ledger"**

Jazyl's dashboard is the ledger at a clinic's front desk: a tidy, familiar console where tomorrow's schedule, the replies to reminders and the tenge saved are written plainly and can be acted on at once. It works in Operate mode. People open it to do a job (confirm a visit, hand out a login, link WhatsApp), not to browse, so every surface is a work surface. The craft target is the category standard set by Linear and Stripe: quiet chrome, exact alignment, tabular figures and hairline borders, finished in Jazyl's binding brand colours and fonts.

The character is calm and restrained. A milk ground carries white panels edged by warm grey hairlines. The deep teal shell holds navigation and identity so the work area can stay neutral. Amber is rare: it marks the one commit a screen exists for. Status is said with small tinted chips, never with colour fields. The system has no decorative charts, vanity tiles, gradients or illustration. Numbers sit in a single divided strip, and depth stops at a whisper of shadow.

Density is moderate: 40px controls on desktop, 20px panel padding, and lists that turn into one card per row on phones so an owner between patients can tap reliably.

**Key Characteristics:**
- Milk ground, white panels, warm grey hairlines; light theme only.
- Deep teal sidebar shell with a white active pill; the work area stays neutral.
- Exactly one amber commit per screen, always with graphite text.
- Manrope for headings and figures, Inter for everything else, both with Cyrillic and Latin Extended subsets.
- Status chips in five tints (neutral, teal, amber, red, green) with a dot.
- Tables from md up, card lists below md.
- No charts, tiles, gradients or decorative imagery.

## Colors

A warm neutral paper palette with one deep teal voice for structure and one amber voice for the commit.

### Primary
- **Yasawi Dome Teal** (brand): the sidebar and phone top bar, the focus outline, the input caret, checkbox accents, links such as phone numbers, the emphasised figure in the stat strip, and the fill of the teal button used for setup and security actions. Its tints are a soft wash (brand-soft) for teal chips and text selection, a line (brand-line) for teal chip rings, and a pale ink (brand-ink) for text on the teal shell. Its deep step (brand-strong) is the hover state and the top bar's bottom border.

### Secondary
- **Warm Commit Amber** (action): only the primary button and the unread badge in the sidebar. Hover darkens it to action-hover. The soft step (action-soft) with the brown ink (action-ink) forms the amber chip and the amber notice, for attention states such as "Просит перенос", "Новая" or "Пилот".

### Neutral
- **Reception Milk** (milk): the page ground and the sticky save bar.
- **Sheet White** (surface): panels, table bodies, inputs, secondary buttons, popovers.
- **Sunk Linen** (sunk): neutral chip fill, neutral notice, hover fill for secondary and ghost buttons, disabled inputs, the day headers in phone card lists (at 60%).
- **Hairline** (line): panel borders, row dividers, the stat strip's 1px gaps, skeleton bars (at 70%).
- **Firm Hairline** (line-strong): input and button borders, the scrollbar thumb.
- **Graphite** (ink): headings, body text, and text on amber.
- **Graphite Mid** (ink-2): descriptions, table headers, secondary labels.
- **Graphite Quiet** (ink-3): hints, placeholders, stat hints, request codes.

### Status
- **Confirmed Green** (ok, ok-soft): confirmed or paying states, success notices.
- **Alarm Red** (bad, bad-soft): cancellations, no-shows, failures, danger buttons, error notices, invalid fields.

Component-level rings and hovers use fixed tints that are not theme tokens: amber ring #efd2a3, red ring #f3c5bf, green ring #bfdfcc, table header and row hover #fbfaf7, primary button border #d98d22 at 40%, input hover border #bfb8a8.

### Named Rules
**The One Commit Rule.** Each screen has exactly one amber primary button: the action the screen exists for ("Добавить запись", "Создать доступ", "Сохранить настройки", "Новая клиника"). Row actions, dialog submits and secondary steps are secondary buttons. If a second amber button appears, one of them is wrong.

**The Graphite on Amber Rule.** Amber always carries graphite (ink) text, including the sidebar badge. White text on amber is forbidden by the brand and fails contrast.

**The Teal for Trust Rule.** The filled teal button is for setup and security steps: linking WhatsApp, starting and confirming two-factor protection. It never competes with the amber commit.

**The Quiet Shell Rule.** Colour lives in the shell, the one commit and the chips. The work area is milk, white and graphite.

## Typography

**Display Font:** Manrope (with Inter, ui-sans-serif)
**Body Font:** Inter (with ui-sans-serif, system-ui), features cv11 and ss01
**Label/Mono Font:** the system monospace stack for keys, codes, temporary passwords and request IDs

**Character:** Manrope's geometric warmth gives headings and figures a friendly, confident voice. Inter keeps labels and tables neutral and readable. Both are loaded with the latin, latin-ext and cyrillic subsets so Russian and Kazakh letters (Қ, Ә, Ұ) render in the same faces.

### Hierarchy
- **Display** (Manrope 700, 26px, tight leading): the page title in the page header.
- **Headline** (Manrope 700, 24px): standalone pages outside the shell, such as forced password change and 2FA setup (20px on login and not-found).
- **Title** (Manrope 600, 15px): panel titles and empty state titles. The error view uses 18px.
- **Body** (Inter 400, 14px): table cells, form values, notices, descriptions under page titles.
- **Body small** (Inter 400 or 500, 13px): panel descriptions, field labels (500), small buttons, phone day headers.
- **Label** (Inter 500, 12px): table headers, chips, hints.
- **Figure** (Manrope 700, 26px, -0.02em, tabular): the values in the stat strip. Appointment times on phones use Manrope 700 at 15px.

### Named Rules
**The Tabular Figures Rule.** Every number a person compares (times, counts, tenge, phone numbers, codes) uses tabular figures. Tables get them by default; elsewhere numbers take the `num` class.

**The Sixteen on Touch Rule.** On coarse pointers inputs, selects and textareas render at 16px so iOS does not zoom.

## Layout

The shell is a fixed 240px teal sidebar from lg (1024px) up. The main column is centred with a maximum width of 1240px and padding of 16px on phones, 24px from sm and 32px from lg (24px top on phones, 32px from lg). Below lg a 56px sticky teal top bar shows the logo, the current clinic name and a menu button that opens the drawer.

Each page begins with the page header: title and description on the left, actions aligned to the bottom right, stacking on phones, then 24px of space. Sections follow in 24px steps. Panels use 20px padding (16px by 20px for headers). Table cells use 16px horizontal padding with 20px at the outer edges, 12px vertical in the body and 10px in the header. The clinic overview pairs a wide schedule panel with a narrower leads panel.

**The Phone List Rule.** Below md (768px) every list renders as cards: one item per row, divided by hairlines, actions laid out at full width or below the content, grouped under milk-linen day headers when time matters. From md up the same data renders as a table with a 640px minimum width inside a horizontal scroller. Both renderings come from the same data in the same component.

## Elevation & Depth

The system is nearly flat. Depth comes from tone (milk ground, white panels, linen wells) and hairline borders. Shadows are kept faint and only separate layers.

### Shadow Vocabulary
- **Panel** (`box-shadow: 0 1px 2px rgb(31 35 40 / 0.05), 0 1px 1px rgb(31 35 40 / 0.03)`): panels, the stat strip, and every filled or secondary button. It is barely visible and adds weight without adding a lift.
- **Pop** (`box-shadow: 0 8px 24px -6px rgb(31 35 40 / 0.18), 0 2px 6px rgb(31 35 40 / 0.06)`): only for layers above the page, namely the password reset popover and the phone drawer.

### Named Rules
**The Two Shadows Rule.** Use Panel or Pop and nothing else. There are no hover lifts, glows or coloured shadows.

## Shapes

Corners are gently rounded in two steps: 8px (control) for buttons, inputs, notices and navigation items, and 12px (panel) for panels, the stat strip, popovers, the error view and chat bubbles. Chips, badges and avatars are full pills or circles. Skeleton bars use 6px. Borders are 1px hairlines everywhere. Panels clip their contents so table rows and strip cells meet the rounded edge cleanly.

## Components

### Buttons
Quiet and solid: medium weight, a 1px press-down on active, 150ms colour transitions.
- **Shape:** control radius (8px); md is 40px tall with 16px padding at 14px, sm is 32px tall with 12px padding at 13px.
- **Primary:** amber with graphite text, a faint darker amber border and the Panel shadow. One per screen (The One Commit Rule).
- **Brand:** teal with white text and a brand-strong border, for setup and security (The Teal for Trust Rule).
- **Secondary:** white with a firm hairline border; hover fills with linen. This is the default variant and the only one used for row actions ("Подтвердил", "Новый пароль") and dialog submits.
- **Ghost:** no fill or border, ink-2 text; hover shows a linen fill.
- **Danger:** white with a red label and firm hairline border; hover fills with bad-soft. Used for "Отмена", "Отключить" and sign-out.
- **Pending:** submit buttons disable at 55% opacity, show a spinning loader icon and swap the label for a present-participle text ("Сохраняем…").
- **Focus:** the global 2px teal outline with 2px offset; inside the teal shell the outline becomes amber.

### Chips
- **Style:** 24px pill, 12px medium text, 10px horizontal padding, a soft tint fill with a 1px inset ring of the same hue and a 6px dot in the current colour.
- **Tones:** neutral (waiting, paused), teal (completed visit), amber (needs attention: reschedule, new lead, pilot, waiting for QR), red (cancelled, no-show, disconnected), green (confirmed, paying, connected). Chips label status only; they are not buttons.

### Panels
- **Corner Style:** 12px.
- **Background:** white on milk.
- **Shadow Strategy:** Panel.
- **Border:** 1px line.
- **Internal Padding:** 20px. An optional header row carries a 15px title, a 13px description and right-aligned actions above a hairline. Flush panels hold tables or lists edge to edge.

### Stat Strip
One row of figures, never a grid of tiles. A single rounded panel is divided into cells by 1px gaps that show the line colour, auto-fitting at a minimum of 150px each. Each cell has a 13px ink-2 label, a 26px Manrope figure and an optional 12px ink-3 hint. One figure per strip may be teal to mark the headline number (saved tenge or MRR).

### Tables
14px body with tabular figures. The header row is #fbfaf7 with 12px medium ink-2 labels. Rows are divided by hairlines, the last row has none, and hover tints a row with #fbfaf7. Action columns align right and hold secondary or danger sm buttons. Group rows such as "Сегодня" and "Завтра" sit on a linen band.

### Inputs / Fields
- **Style:** 40px tall, white, firm hairline border, 8px radius, 12px padding, a faint inset top shadow; the caret is teal.
- **Focus:** the border turns teal with a 3px teal ring at 15%.
- **Error / Disabled:** aria-invalid gives a red border and a red ring; disabled fields are linen with ink-3 text.
- **Field:** a 13px medium label above, a 12px ink-3 hint below, 6px gaps. Checkboxes use the teal accent with the label and hint beside them.
- **Codes:** one-time codes use a 48px centred field with Manrope 24px and wide tracking.

### Navigation
- **Sidebar:** full height teal. Logo row 64px, then an optional back link and a context card (a translucent white panel with the clinic name and address), then the nav list. Items are 40px, 14px, with an 18px line icon. The active item is a white pill with teal semibold text and the Panel shadow. Hover is white at 8%. An amber badge with graphite text shows unread counts. The footer holds the account link (initial avatar, name, role) and sign-out, above a white 10% hairline.
- **Phone:** a 56px sticky teal top bar. The menu button opens a left drawer (up to 18rem or 86vw) that reuses the same sidebar, slides in over 300ms with the out easing, sits over a graphite backdrop at 40% and carries the Pop shadow. Escape, the backdrop or any link closes it, and the page behind stops scrolling.

### Password Reset Popover
The per-user password reset is a native top-layer popover (`popover="auto"`) with a dialog role, opened from the row's secondary "Новый пароль" button. It is centred, up to 24rem wide (never wider than the viewport minus 32px), white with a hairline, 12px radius, 20px padding and the Pop shadow, over a graphite backdrop at 30%. The header has a 16px semibold title, the user's email in 12px ink-2 and an icon close button. The body is an inline action form with the temporary password field and a secondary submit, because the screen's amber commit belongs to the page.

### Sticky Save Bar
Long settings forms end in a bar fixed to the bottom of the viewport. It is solid milk (never translucent or blurred) with a hairline top border and 12px vertical padding, and it holds the screen's single amber "Сохранить настройки" commit.

### Empty States, Notices, Loading and Errors
- **Empty:** inside the panel that would hold the list: a 15px Manrope title, a short explanation of what to do next (at most 28rem wide) and at most one action. It is left-aligned on phones and centred from sm.
- **Notice:** an 8px rounded tinted box with 16px by 12px padding in neutral, amber, red or green, used for form results (announced politely) and inline warnings.
- **Loading:** route skeletons keep the page's shape (title bar, optional strip of four cells, rows with time, text and chip placeholders) in pulsing line-tinted bars, marked as busy.
- **Error:** a white panel with a role of alert, an 18px title, a plain explanation that nothing was lost, the request code in monospace when present, and a secondary "Повторить" button.

## Do's and Don'ts

### Do:
- **Do** give every screen exactly one amber primary commit with graphite text, and make row actions secondary or danger sm buttons.
- **Do** use the teal brand button for WhatsApp linking and two-factor setup.
- **Do** show numbers in one divided stat strip with tabular figures in Manrope 700 at 26px.
- **Do** render lists as cards below 768px and as tables from 768px up.
- **Do** use Panel (`0 1px 2px rgb(31 35 40 / 0.05), 0 1px 1px rgb(31 35 40 / 0.03)`) for resting surfaces and Pop only for the popover and drawer.
- **Do** say status with the five chip tones and pair every chip colour with a text label.
- **Do** show empty states, failures and loading plainly in the place the data would appear.
- **Do** keep sticky bars solid milk with a hairline border.

### Don't:
- **Don't** put white text on amber, or use amber for anything other than the commit, the unread badge and attention tints.
- **Don't** place two amber buttons on one screen.
- **Don't** add decorative charts, KPI tiles, card grids of numbers, gradients or illustration.
- **Don't** add a dark theme; the dashboard is used at reception desks and on phones in daylight.
- **Don't** use shadows other than Panel and Pop, or add hover lifts and glows.
- **Don't** use teeth, smiles, syringes or crosses as imagery or icons; the logo's checkmark is the only mark.
- **Don't** use a translucent or blurred sticky bar.
