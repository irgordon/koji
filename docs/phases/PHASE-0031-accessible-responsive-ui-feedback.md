# PHASE-0031: Accessible Responsive UI Feedback

## Objective

Improve Koji's frontend feedback, accessibility, and responsive behavior without changing backend security or privileged behavior.

## What Changed

- Added a reusable toast system with success, error, warning, and info messages.
- Added live-region notifications for service-control jobs, job decisions, permission/session failures, and API failures.
- Added clearer inline notices for allowlists, redaction policy, job approval, and agent-bound service control.
- Added last-updated and stale-data indicators for Overview, Jobs, and Activity surfaces.
- Added non-color-only status badge labels.
- Improved keyboard focus states, tap targets, tooltip focus behavior, and mobile layout.
- Slowed polling for higher-cost views:
  - Overview host metrics: moderate polling
  - Services: only while Services is active
  - Processes: slower and only while Processes is active
  - Jobs: moderate and only while Jobs is active
  - Activity: slower and only while Activity is active

## Accessibility Checks

- Semantic landmarks are present for navigation, header, and main content.
- Navigation buttons expose current-page state with `aria-current`.
- Toasts use live regions and error toasts use alert semantics.
- Tooltips are keyboard focusable and use `aria-describedby`.
- Loading, inline notice, and error states expose status or alert semantics.
- Focus-visible styles are explicit for buttons, inputs, navigation items, and tooltips.
- Status badges include text labels, not color alone.
- Tap targets are sized for mobile interaction.
- Reduced-motion preferences are respected in CSS.

## Responsive Checks

Checked target widths:

- 375px mobile
- 768px tablet
- 1024px desktop

Expected behavior:

- Navigation wraps into usable groups on small screens.
- Cards stack cleanly.
- Tables remain horizontally scrollable instead of breaking the layout.
- Job approval and rejection controls wrap and remain reachable.
- Toasts anchor to the viewport without covering primary navigation.
- Tooltips can be focused without hover.

## Known Limitations

- The production login/bootstrap experience is still served by the backend session surfaces rather than a dedicated React login screen.
- Browser checks with the frontend-only Vite server show expected API failure feedback because the Go API is not running behind the dev server.
- Tooltips remain lightweight CSS/HTML help text; they are not a fully managed popover system.

## Validation

- `npm run build`
- `rg -n ": any|as any" web/src`
- `rg -n "catch\\s*\\([^)]*\\)\\s*\\{\\s*\\}" web/src`
- `gofmt -w cmd internal`
- `GOCACHE=/tmp/koji-go-cache go test ./...`
- `git diff --check`
- `rg -n "exec\\.Command|CommandContext" internal/http internal/agent internal/system`
- `rg -n "systemctl" internal/http internal/agent internal/system internal/jobs`
