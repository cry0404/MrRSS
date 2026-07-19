# Open Issue Review

Date: 2026-06-09

Scope: current open issues in `WCY-dt/MrRSS`, reviewed against the current `release/v1.3.24` branch.

## Can Close After This Version Ships

These issues have direct code changes in the current version and automated verification coverage or successful build coverage:

- #920 LibreTranslate custom API template error
- #902 Turkish language support for translation options
- #911 LAN Ollama/local AI endpoint TLS/client behavior
- #875 full-text fetch did not use proxy
- #912 newsletter sender filter did not work
- #896 feed management search plus select-all selected hidden rows
- #802 cache size cleanup made old deleted articles reappear
- #669 unread view pagination/list behavior
- #801 option to disable startup update prompt
- #903 article content links open in the system browser
- #619 mark all as read refreshes and disrupts current list/reading context
- #427 copy article link button in the article toolbar

## Can Close After Reporter Confirmation

These are addressed by code-level fixes in this version, but the reported behavior depends on OS window managers, live feeds, external APIs, credentials, or site-specific behavior that cannot be fully reproduced in the current Windows development environment:

- #913, #770, #716, #643, #320 macOS/window restore, dock reopen, and maximize behavior
- #917, #826, #873 old/read articles reappearing or showing in unread views
- #918, #909, #772, #767, #750 translation and AI failures across external providers
- #467, #601, #795, #876, #894 full-text extraction, subscription, encoding, and site-specific feed edge cases

## Low-Complexity Issues Completed In This Pass

- #801
  - Added `update_check_enabled` to the schema-driven settings system.
  - Added a Settings > General > Updates toggle.
  - Startup update checking and the update dialog are skipped when disabled.

- #903
  - Article body links rendered from RSS/full-text HTML now open through `/api/browser/open`.
  - Relative links are resolved against the article URL.
  - In-page anchors and non-http(s) links are left alone.

- #619
  - Mark-all-read no longer refetches the article list.
  - Current in-memory articles are marked read in place.
  - Unread/filter counts are refreshed without clearing the current list.

- #427
  - Added a copy-link button to the article toolbar.
  - Reuses the existing native clipboard utility and toast feedback.

## Reviewed But Not Pulled Into This Pass

- #910 RSS original summary as a summary provider
  - Feasible, but broader than a small UI toggle because the current article list/detail queries and summary handler do not consistently expose article `description`.
  - Recommended implementation: add a `rss` summary provider, load `articles.description` in article queries, return/copy-render the RSS description without overwriting AI/local summary cache unless explicitly desired.

- #555 single article refresh
  - Feasible but touches article content refresh, unread-mode selection behavior, and cache invalidation. Better handled as a focused feature.

- #594 single category refresh
  - Feasible but touches task scheduling and refresh progress semantics. Better handled separately.

- #828 per-feed cookies
  - Not a small change. It affects feed model/schema, HTTP client construction, security handling, OPML/export behavior, and UI.

- #895 Windows x86 installer availability
  - Release packaging/pipeline decision, not an app code fix.
