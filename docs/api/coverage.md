# Requirement Coverage And Gaps

## Coverage Snapshot

| Phase | Status | Notes |
| --- | --- | --- |
| Phase 1, anonymous MVP | Passed | Anonymous session flow passed API testing, including wizard saves, recommendations, PDF generation, and PDF download |
| Phase 2, auth and project work | Partial | Register, login, refresh, logout, and protected project routes exist; local tests pass, but deployed registration should be retested with production Redis/Postgres settings |
| Phase 3, session conversion | Passed locally | Conversion creates a project from the session, copies an existing PDF reference, and deletes the source session |
| Phase 4, refinement and optimization | Partial | Activity tracking is implemented, but performance, centralized monitoring, and PDF format requirements were not fully verified |

## Observed Test Results

- `GET /health` returned `200`
- `GET /ready` returned `200` with Postgres and Redis marked `ok`
- `GET /api/v1/ping` returned `200`
- `GET /` returned `404`
- Invalid registration data returned `400`
- Valid registration previously returned `500` in the deployed test environment; local auth tests now pass after refresh-token test wiring was fixed
- Login for a missing user returned `401`
- Protected routes without a token returned `401`
- Browser UI testing was blocked because Playwright Chrome was missing locally

## Known Gaps

- No canonical frontend is mounted at the root URL, so the deployment behaves like an API only backend
- Project listing has no pagination
- PDF generation still returns a public storage URL; authenticated project download and anonymous session download endpoints are also available
- Production object storage and email are not enabled on the test deployment
- PDF/A and PDF 1.7 output were not verified
- Performance and external monitoring requirements were not fully verified

## Current Contract Notes

- `GET /files/pdfs/*` serves public PDF bytes
- `activity_events` stores API request and key user-flow activity records
- The error envelope is always `{ "error": { "code", "message", "details" } }`
- `OPTIONS` requests return `204 No Content`; CORS allow headers are added only for allowed origins
