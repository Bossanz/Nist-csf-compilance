# Same-origin API proxy design

## Problem

The browser bundle currently calls `http://localhost:8080` directly. This creates a cross-origin browser-to-API dependency and produces a persistent `Failed to fetch` state when the API is restarting or temporarily unavailable, even though the Next.js application is available on port 3000.

## Design

The browser will call relative `/api/...` paths on the Next.js origin. Next.js will rewrite those requests to the Go API using an internal server-side base URL. Under Docker Compose, that target is `http://api:8080`, which uses the Compose service network and does not expose container addressing to browser code.

The Go API and its routes remain unchanged. Port 8080 may remain published for local diagnostics, but normal browser traffic will use port 3000 only.

## Configuration

- `web/src/lib/api.ts` uses an empty public base by default so requests remain same-origin.
- `web/next.config.ts` defines an `/api/:path*` rewrite to `${API_INTERNAL_URL}/api/:path*`.
- Docker Compose passes `API_INTERNAL_URL=http://api:8080` to the web container and removes the browser-facing `NEXT_PUBLIC_API_BASE_URL` setting.
- Local development defaults the internal API target to `http://localhost:8080`.

## Error handling

Existing API error handling remains unchanged. HTTP errors continue to show the API message. A genuinely unavailable proxy target still produces an error, but browser requests no longer depend on cross-origin access or the host-visible API address.

## Testing

- Add a focused test proving browser API requests use relative `/api/...` URLs.
- Run the complete frontend test suite and production build.
- Rebuild and restart Docker Compose.
- Verify the web page, health endpoint, project listing, project creation/opening, and browser console from the in-app browser.

## Scope

No retry framework, new backend endpoint, authentication change, or unrelated refactor is included.
