# PHASE-0032: Frontend Test Harness and Accessibility Regression Suite

## Objective

Add a lightweight frontend test layer that protects the Phase 31 accessibility, feedback, and responsive UI work from regression.

## Scope

- Added Vitest with jsdom.
- Added Testing Library, jest-dom matchers, and user-event.
- Added `npm run test` and `npm run test:watch`.
- Added frontend setup under `web/src/test/setup.ts`.
- Added component tests for:
  - Toast live region, dismiss, and auto-dismiss behavior
  - Tooltip role, focusability, and `aria-describedby`
  - StatusBadge non-color-only labels
  - ErrorBanner safe plain-text rendering
  - API error normalization
  - Jobs page status and approval controls
  - Activity read model rendering
  - Observability cards
  - Mobile-width shell smoke behavior
- Updated release CI to run frontend tests before the frontend build.

## Boundaries

- No backend behavior changed.
- No API contracts changed.
- No browser automation added.
- No screenshot or visual snapshot testing added.
- No Playwright, Cypress, Selenium, Puppeteer, or pixel comparison testing added.

## Accessibility Coverage

The tests assert:

- Toast region has `aria-live`.
- Error feedback uses alert semantics.
- Tooltips expose `role="tooltip"` and `aria-describedby`.
- Tooltip trigger is keyboard focusable.
- Status labels are not color-only.
- Job decision inputs have accessible labels and descriptions.

## Known Limitations

- This is a component and jsdom regression layer, not an end-to-end browser suite.
- Responsive checks verify mounted controls and narrow viewport assumptions, not computed pixel-perfect layout.
- Full login/bootstrap page behavior remains a future target once the React app owns those session surfaces directly.

## Validation

- `npm run test`
- `npm run build`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
- `git diff --check`
- `rg -n "aria-live|role=\"tooltip\"|aria-describedby" web/src`
