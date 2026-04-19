# Market Lens Change Log

This file records major prompts/changes after completion.

Each entry should capture:

1. Task in one sentence.
2. Bullet points of things done.
3. Corrections or steering required.

Commit workflow:

- Only commit after the user gives an explicit signal.
- Inspect status and diffs before staging.
- Do not blindly commit everything.
- Bundle related changes into sensible commits.
- Push only after the grouped commits are created and verified.

Token discipline workflow:

- Reference `TOKEN_DISCIPLINE.md` before future work.
- Use targeted file inspection.
- Keep command output compact.
- Avoid full JSON/log dumps.
- Keep progress updates and final summaries concise unless exhaustive detail is requested.

## 1. Initial V1 Build

### Task In One Sentence

Build the first local Market Lens MVP as a React + Go + SQLite stock-analysis dashboard with Docker persistence, snapshots, and an analysis timeline.

### Things Done

- Created the project root at `market-lens`.
- Built a React frontend with a dark dashboard UI.
- Built a Go backend API.
- Added SQLite persistence.
- Added Docker Compose with separate `api` and `web` services.
- Mounted `./data:/data` so the SQLite database persists outside containers.
- Added ticker creation and watchlist storage.
- Added quote, chart, news, SEC filing, sentiment, and summary analysis behavior.
- Added worker-event tracking.
- Added an analysis timeline.
- Added manual SQLite snapshots using `VACUUM INTO`.
- Added mock fallback data when API keys are missing.
- Added optional Alpha Vantage quote/news support.
- Added basic SEC filing lookup for common symbols.
- Added README documentation for the initial v1 baseline.
- Built and verified the frontend.
- Built and verified the Go backend.
- Built and verified Docker once Docker was running.

### Corrections Or Steering Required

- User clarified the stack should be React UI, Go backend, and SQLite database.
- User clarified there should be no validation for now.
- User required snapshots so data would not be lost.
- User required Dockerization with the database volume outside the app runtime.
- User required all analyses to be tracked and available as a timeline.
- Docker could not build initially because Docker daemon was not running; retried after user started Docker.
- Adding `AAPL` failed with a 404 because the frontend called `/api/api/stocks`; fixed the API base path so it calls `/api/stocks`.

---

## 2. Modular One-View Workflow

### Task In One Sentence

Simplify the main page into a single scan-friendly view and add a threaded-style analysis workflow with visible loading/progress states.

### Things Done

- Followed the project workflow by documenting the intended change in the README before implementation.
- Added async workflow APIs:
  - `POST /api/runs`
  - `GET /api/runs/{id}`
- Added in-memory workflow run tracking.
- Added stage-by-stage workflow status for:
  - quote
  - news
  - SEC filings
  - sentiment
  - final summary
- Kept the older synchronous stock endpoints available.
- Refactored backend code out of one large file into focused modules:
  - `main.go`
  - `types.go`
  - `routes.go`
  - `db.go`
  - `workflow.go`
  - `providers.go`
  - `scoring.go`
  - `utils.go`
- Refactored frontend code into focused modules:
  - `App.jsx`
  - `api.js`
  - `components.jsx`
  - `format.js`
  - `main.jsx`
  - `styles.css`
- Reworked the UI into a simpler one-view dashboard.
- Added a command bar for running analysis.
- Added a workflow rail showing processing stages.
- Added spinner/loading indicators while analysis is running.
- Added polling from the frontend to the backend workflow endpoint.
- Updated README with the modularity and workflow implementation record.
- Rebuilt and verified frontend, backend, and Docker.

### Corrections Or Steering Required

- User clarified the UI should avoid clutter and show everything in one view.
- User clarified the workflow should run through the analysis steps and then give a final summary.
- User clarified the UI should show a loading icon or processing indicator while work is running.
- User interrupted/observed that the task may not have completed; inspected the repo, found the CSS and final verification were unfinished, then completed them.
- User clarified code should stay modular and not be dumped into one file; refactored backend and frontend into focused modules before finishing the workflow.
