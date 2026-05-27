# Activity Tracking

The backend records API and user-flow activity in the `activity_events` table.

## Storage

Migration `000004_activity_events.sql` creates:

- `event_type`
- `user_id`
- `session_id`
- `project_id`
- `method`
- `path`
- `status_code`
- `duration_ms`
- `metadata`
- `created_at`

The middleware records every request without changing the response. If activity persistence fails, the request still completes and the failure is logged.

## Event Types

Main event types include:

- `api_request`
- `api_request_failed`
- `server_error`
- `user_registered`
- `user_login`
- `user_logout`
- `token_refreshed`
- `session_created`
- `session_converted`
- `project_created`
- `project_updated`
- `project_deleted`
- `wizard_step_saved`
- `recommendation_requested`
- `pdf_generated`
- `pdf_downloaded`

## Identity Fields

For authenticated calls, `user_id` is derived from the bearer token when available. For anonymous flows, `session_id` is extracted from the route path. Project routes populate `project_id` from the route path when present.
