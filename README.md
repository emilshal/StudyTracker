# Study Tracker

Study Tracker is a starter project for logging study sessions, visualising progress, and building consistent study habits. The backend is written in Go, while the frontend uses HTML/CSS/JavaScript.

## Project layout

```
StudyProject/
├── backend/                 # Go API service
│   ├── cmd/server/main.go   # Application entry point
│   ├── go.mod               # Go module definition
│   └── internal/            # Private packages for business logic
│       ├── router/          # HTTP router wiring
│       └── study/           # Study domain (handlers, services, models)
└── frontend/                # Static web client
    ├── index.html
    └── assets/
        ├── css/styles.css
        └── js/app.js
```

## Running the backend

```bash
cd backend
go run ./cmd/server
```

This serves the API on `http://localhost:8080` and also exposes the static frontend from the root path for quick prototyping. By default the backend persists data to `backend/data/studytracker.db` using SQLite. Override with `DATABASE_URL`, for example:

```bash
DATABASE_URL="file:data/dev.db?_pragma=foreign_keys(ON)" go run ./cmd/server
```

Migrations are applied automatically on startup from `internal/platform/database/migrations`.

### Environment variables

- `DATABASE_URL` – SQLite DSN (default `file:data/studytracker.db?_pragma=foreign_keys(ON)`).
- `SESSION_TTL` – optional duration for session lifetime (default `24h`).
- `FRONTEND_URL` – URL to redirect after OAuth callback (default `/`).
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URL` – optional; provide to enable Google Sign-In. The redirect URL can point to `https://<host>/api/auth/google/callback` or the alias `https://<host>/oauth/callback`.

### Authentication flow

Users can either register with email/password or continue with Google. Successful logins receive a server-backed session stored in an HTTP-only cookie. The SPA keeps unauthenticated visitors on the Auth view until they sign in; once authenticated, dashboard, history, log, and trends views become available.

## Local development roadmap

- Harden persistence (indexes, backups) and add a migration CLI.
- Add authentication (session cookies or JWT) to support multiple users.
- Build filtering and pagination for study sessions.
- Enhance study session logging with timers, editing, and richer validation.
- Expose richer analytics (daily/weekly charts, streaks, goal tracking).
- Add background jobs for reminders, notifications, or spaced repetition hints.
- Write automated tests for services and handlers.

## Currently implemented

- Log study sessions with subject selection (or free entry), start/end times, auto-calculated duration, notes, and reflections.
- View recent sessions and aggregated totals in the prototype UI.
- Manage a curated subject catalogue (add, edit, delete) with colour tags.
- Edit or delete logged study sessions directly from the UI.
- Interactive dashboard with daily/weekly/monthly totals, streak tracking, and charts for subjects and 14-day trends.

## Feature ideas

1. **Session logging** – Track subject, duration, notes, and mood for each study block.
2. **Progress dashboards** – Visualise total time, per-subject breakdown, and streaks.
3. **Goal setting** – Define weekly targets with progress indicators and alerts.
4. **Study plans** – Create reusable study templates and checklists.
5. **Insights & recommendations** – Surface trends, suggest subjects needing attention, or recommend Pomodoro cycles.
6. **Collaboration** – Share progress with friends, mentors, or accountability groups.
7. **Reminders & notifications** – Send nudges via email/push when falling behind goals.
8. **Data export** – Provide CSV/JSON exports and calendar integration.
