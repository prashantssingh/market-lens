# Market Lens

Market Lens is a personal local stock-analysis dashboard. v1 is intentionally simple: a React UI, a Go backend, SQLite persistence, Docker support, and a timeline of every analysis run.

The app is not a trading terminal. It does not place orders, connect to a broker, or provide financial advice. It is a local research dashboard that helps answer: "Is this stock worth investigating further right now, and why?"

## V1 Implementation Record

This is the implementation baseline created for v1.

### Product Shape

- Built a simple dark dashboard inspired by the provided trading-terminal screenshot, but with fewer panels and no trade execution controls.
- Added a left watchlist, central selected-stock workspace, right signal summary, lower intelligence panels, and an analysis timeline.
- Added a manual snapshot action so the current SQLite database can be copied to a timestamped snapshot file.
- Kept v1 focused on a signal dashboard instead of a full research terminal.

### Stack

- Frontend: React + Vite.
- Backend: Go HTTP server.
- Database: SQLite through `modernc.org/sqlite`.
- Runtime: Docker Compose with separate `api` and `web` services.
- Web serving/proxy: Nginx in the frontend container.

### Backend

The Go backend lives in `backend/`.

Implemented API behavior:

- `GET /api/health`: health check.
- `GET /api/state`: returns stocks, analysis timeline, and snapshots.
- `GET /api/stocks`: returns saved stocks.
- `POST /api/stocks`: adds a stock and immediately runs an analysis.
- `POST /api/stocks/{symbol}/analyze`: runs a new analysis for an existing or new symbol.
- `GET /api/snapshots`: lists snapshots.
- `POST /api/snapshots`: creates a timestamped SQLite snapshot.

Implemented persistence tables:

- `stocks`: local watchlist.
- `analyses`: full analysis payloads for the timeline.
- `worker_events`: quote, news, SEC filing, sentiment, and analysis worker status events.
- `snapshots`: records created snapshot files.

Implemented analysis workers:

- Quote worker.
- News worker.
- SEC filing worker.
- Sentiment worker.
- Analysis worker.

Each analysis run stores a full payload containing:

- Quote.
- Chart points.
- News/catalyst items.
- SEC filing items.
- Worker events.
- Signal summary.
- Confidence.
- Reasons.
- Warnings.

### Data Providers

v1 uses free/free-tier-friendly behavior.

- If `ALPHA_VANTAGE_API_KEY` is set, the backend attempts Alpha Vantage quote/news calls.
- If no API key is present, the app uses deterministic mock quote/news data so the UI and pipeline still work.
- SEC filing lookup is implemented for common symbols using SEC submissions data, with conservative fallback behavior.
- Options and institutional analysis are explicitly deferred from v1.

Known SEC symbol coverage in v1:

- AAPL
- MSFT
- NVDA
- TSLA
- AMZN
- GOOGL
- META

### Frontend

The React frontend lives in `frontend/`.

Implemented UI sections:

- Top app bar with snapshot action.
- Watchlist with add-ticker form.
- Selected stock header.
- Quote strip.
- Simple SVG price chart.
- Signal summary with reasons and warnings.
- News/catalyst feed.
- SEC filings feed.
- Worker status feed.
- Analysis timeline.
- Snapshot summary row.

The frontend calls relative `/api/...` routes in Docker so Nginx can proxy requests to the Go API.

### Docker

Docker Compose lives at `docker-compose.yml`.

Services:

- `api`: Go backend on container port `8080`, published as host port `8080`.
- `web`: Nginx-served React build on container port `80`, published as host port `3000`.

Persistence:

- The backend uses `DATA_DIR=/data`.
- Compose bind-mounts `./data:/data`.
- This keeps the SQLite DB outside the container so app restarts pick up the existing data.

Persistent files:

```text
./data/market-lens.db
./data/snapshots/
```

### Snapshot Behavior

Snapshots are created by calling `POST /api/snapshots` or pressing the Snapshot button in the UI.

The backend uses SQLite `VACUUM INTO` to copy the active DB into:

```text
./data/snapshots/market-lens-YYYYMMDD-HHMMSS.db
```

One smoke-test snapshot was created during v1 verification:

```text
./data/snapshots/market-lens-20260418-190506.db
```

### Bug Fixed During V1

After Docker startup, adding `AAPL` failed with a 404.

Root cause:

- The built frontend used `VITE_API_BASE=/api`.
- Frontend requests already used paths like `/api/stocks`.
- Together, that produced `/api/api/stocks`.
- Nginx logs confirmed requests like `POST /api/api/stocks HTTP/1.1" 404`.

Fix:

- Changed frontend API base to default to an empty string.
- Changed Docker build arg `VITE_API_BASE` to an empty string.
- Rebuilt and restarted the web container.

Verified fix:

- `POST /api/stocks` through the web container returns `200`.
- Nginx logs now show `POST /api/stocks HTTP/1.1" 200`.

Files changed for the fix:

- `frontend/src/main.jsx`
- `frontend/Dockerfile`
- `docker-compose.yml`

### One-View Workflow Update

The next step after the initial MVP simplified the app and made the analysis flow more explicit.

Implemented:

- Reworked the frontend into a single scan-friendly dashboard view.
- Added a visible workflow rail that shows quote, news, SEC filings, sentiment, and final summary stages.
- Added a spinner/loading state while the threaded analysis workflow is processing.
- Added async workflow APIs:
  - `POST /api/runs`: starts an analysis workflow and returns a run id.
  - `GET /api/runs/{id}`: polls workflow status until complete or failed.
- Kept the old synchronous stock endpoints available for compatibility.
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

Verification:

- `go test ./...` passed.
- `npm run build` passed.
- Docker images rebuilt successfully.
- Docker stack restarted successfully.
- Container health check passed.
- `POST /api/runs` through the web proxy returned a running workflow.
- Polling `GET /api/runs/{id}` returned a completed workflow with a saved timeline analysis.

## Run With Docker

From the project root:

```bash
docker compose up --build
```

Open:

```text
http://localhost:3000
```

Stop:

```bash
docker compose down
```

Check containers:

```bash
docker compose ps
```

View logs:

```bash
docker compose logs --tail=100
```

## Run Locally

Backend:

```bash
cd backend
DATA_DIR=../data PORT=8080 go run .
```

Frontend:

```bash
cd frontend
npm install
npm run dev -- --port 3000
```

Open:

```text
http://localhost:3000
```

## Optional Free API Key

Market Lens works without API keys by using mock fallback data.

To enable Alpha Vantage quote/news attempts, create `.env` or export:

```bash
ALPHA_VANTAGE_API_KEY=your_key_here
```

Docker Compose passes this value into the API container if it is present.

## Verification Completed

Commands/checks completed during v1:

- `go mod tidy`
- `go test ./...`
- `npm install`
- `npm run build`
- One-shot API smoke test for `GET /api/health`
- One-shot API smoke test for `POST /api/stocks` with `AAPL`
- One-shot snapshot smoke test for `POST /api/snapshots`
- `docker compose config`
- `docker compose build`
- `docker compose up -d`
- `docker compose ps`
- Docker network check from `web` to `api`
- Proxy check for `POST /api/stocks`

Current verified browser URL:

```text
http://localhost:3000
```

## V1 Boundaries

Included:

- Watchlist.
- Add ticker.
- Quote/chart display.
- News/catalyst display.
- SEC filing display.
- Worker status tracking.
- Analysis timeline.
- SQLite persistence.
- Manual snapshots.
- Dockerized app with external volume.

Deferred:

- Full options chain analysis.
- Institutional ownership analysis.
- Full insider transaction parsing.
- Broker connection.
- Trading execution.
- User accounts/auth.
- Production deployment.
- Complex multi-model orchestration.

## Notes For Next Work

- Project workflow for every future step: plan first, update this README with the planned/accepted change, then implement it in the app.
- Token discipline workflow: before each future task, reference `TOKEN_DISCIPLINE.md` and keep file reads, tool outputs, patches, verification, and summaries compact.
- Next accepted step: simplify the main webpage into one scan-friendly view, add an analysis workflow runner that executes the quote, news, SEC filing, sentiment, and final-summary stages as tracked work, and show a loading/progress indicator while the workflow is processing.
- Implementation constraint for the next step: keep the code modular and avoid dumping backend or frontend logic into one large file. Backend work should be separated by routes, database access, workflow orchestration, providers, scoring, and utilities. Frontend work should be separated into API helpers, app state, and reusable view components.
- Commit workflow: only commit after the user gives an explicit signal; inspect status/diffs first; do not blindly commit everything; group related changes into sensible commits before pushing.
- The data model is intentionally payload-oriented in v1, with full analysis JSON stored in `analyses.payload`.
- This is good enough for the MVP timeline and fast iteration.
- A later version can normalize quote/news/filing tables further if filtering, searching, or analytics become more important.
- The worker model is currently synchronous but tracked as named worker events. This gives us a clean migration path to real background jobs later.
- The app should continue to show explicit freshness/warning states instead of pretending all free-tier data is real time.
