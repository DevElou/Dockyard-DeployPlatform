# Product

## Register

product

## Users

Solo developer (homelab owner) managing personal infrastructure. Uses this daily, alone, on a 13–27" screen, usually at a desk. Knows the stack intimately — no need for hand-holding or progressive onboarding copy. The job: get code deployed, confirm it's healthy, move on.

## Product Purpose

Dockyard is a private deployment platform (Vercel-like) for a homelab ESXi infrastructure. It replaces manual Docker CLI operations with a clean web UI: create a project, trigger a build from GitHub, deploy to a Docker host, monitor status. No auth, no multi-tenant complexity — one operator, full control.

Success looks like: open Dockyard, see that everything is green, close Dockyard. Or: trigger a deploy, watch it go healthy, move on in under two minutes.

## Brand Personality

Calm, clear, confident. A tool that works and doesn't show off. Quiet authority — the way a well-made instrument feels in your hands. Not cute, not corporate, not tryhard.

## Anti-references

- **Grafana / monitoring dashboards** — chart-first layouts, dark slate, LED status palette, metric overload
- **Generic SaaS cream** — rounded purple gradient heroes, pricing tables, startup marketing feel
- **Terminal emulator UIs** — monospaced everything, bright green on black, hacker aesthetic
- **Portainer / Docker Desktop** — icon-heavy, cluttered, container-manager grey

## References

- **Vercel dashboard** — deployment-centric, status badges, crisp hierarchy, white, focused
- **Linear** — fast, great typographic hierarchy, precise spacing, confident defaults

## Design Principles

1. **Status earns first position.** Every screen's primary job is to surface health at a glance. Healthy/failing/building states are never secondary to chrome or navigation.
2. **Earn every pixel.** No decorative sections, no padding for its own sake. Each element carries information or enables action. Whitespace is structure, not filler.
3. **One accent, used deliberately.** A single crisp accent color — never a rainbow of semantic state colors. Green/red status signals serve the data; the accent serves the brand.
4. **Quiet confidence.** No tooltips explaining what buttons do. No empty-state animations. The UI trusts the user knows their infrastructure.
5. **Defaults hide complexity.** Healthcheck paths, deploy strategies, and port configurations have sensible defaults. Complexity is revealed on intent, not upfront.

## Accessibility & Inclusion

WCAG AA as a baseline. Single user, no known specific needs. Reduced motion respected via `prefers-reduced-motion`. Color is never the only signal (status always uses icon + text alongside color).
