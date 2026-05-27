# Activity Tracking

The backend records API and user-flow activity in the `activity_events` table.

## `POST /api/v1/activity/events`

Frontend clients can record UI-level activity that does not naturally hit an existing backend route, such as page views, button clicks, form interactions, and wizard navigation.

| Field | Value |
| --- | --- |
| Auth | Optional bearer token |
| Success | `202` `{ "status": "accepted" }` |
| Errors | `400 INVALID_INPUT`, `400 INVALID_ACTIVITY_EVENT`, `401 UNAUTHORIZED`, `500 ACTIVITY_RECORD_FAILED`, `503 ACTIVITY_RECORDER_UNAVAILABLE` |
| Notes | Without auth, send `sessionId` for anonymous flows. With a bearer token, `user_id` is attached automatically. |

Request:

```json
{
  "type": "button_click",
  "sessionId": "1f9f2b8d-1f0b-4c3c-9e2c-3dbd8f8b2d77",
  "projectId": "4ec4aa78-4ce0-4a77-aad1-5f74b66b1f5b",
  "page": "/wizard/step/2",
  "metadata": {
    "target": "next",
    "component": "WizardFooter"
  }
}
```

The stored event type is prefixed with `frontend_`, so the example above is stored as `frontend_button_click`.

Validation:

- `type` is required, max 64 characters, and may contain letters, numbers, `_`, `-`, or `.`
- `sessionId` and `projectId` are optional and capped at 128 characters
- `page` is optional and capped at 512 characters
- `metadata` is optional and capped at 16 KB after JSON encoding

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
- `frontend_*`

## Identity Fields

For authenticated calls, `user_id` is derived from the bearer token when available. For anonymous flows, `session_id` is extracted from the route path. Project routes populate `project_id` from the route path when present.

For `POST /api/v1/activity/events`, `session_id`, `project_id`, and frontend page path come from the request body, while `user_id` comes from the optional bearer token.
