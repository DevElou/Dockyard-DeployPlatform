---
name: Dockyard
description: Private deployment platform — manage Docker hosts from a single, clear interface.
colors:
  canvas: "oklch(98.5% 0.007 240)"
  foreground: "oklch(12% 0.025 250)"
  primary: "oklch(22% 0.03 250)"
  primary-fg: "oklch(97% 0.004 240)"
  muted: "oklch(95.5% 0.008 240)"
  muted-fg: "oklch(48% 0.018 250)"
  border: "oklch(90% 0.010 240)"
  ring: "oklch(22% 0.03 250)"
  status-healthy: "oklch(52% 0.13 145)"
  status-building: "oklch(55% 0.13 230)"
  status-pending: "oklch(58% 0.12 75)"
  status-failed: "oklch(52% 0.18 28)"
typography:
  display:
    fontFamily: "Inter Variable, Inter, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "-0.02em"
  headline:
    fontFamily: "Inter Variable, Inter, system-ui, sans-serif"
    fontSize: "1.125rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "-0.01em"
  title:
    fontFamily: "Inter Variable, Inter, system-ui, sans-serif"
    fontSize: "0.9375rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "normal"
  body:
    fontFamily: "Inter Variable, Inter, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.6
    letterSpacing: "normal"
  label:
    fontFamily: "Inter Variable, Inter, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "0.01em"
rounded:
  xs: "3px"
  sm: "5px"
  md: "8px"
  lg: "12px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
  2xl: "48px"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-fg}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  button-primary-hover:
    backgroundColor: "oklch(30% 0.03 250)"
    textColor: "{colors.primary-fg}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  button-outline:
    backgroundColor: "transparent"
    textColor: "{colors.foreground}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  button-outline-hover:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  button-ghost:
    backgroundColor: "transparent"
    textColor: "{colors.muted-fg}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  button-ghost-hover:
    backgroundColor: "{colors.muted}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.sm}"
    padding: "7px 14px"
  card:
    backgroundColor: "{colors.canvas}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: "16px"
  input:
    backgroundColor: "transparent"
    textColor: "{colors.foreground}"
    rounded: "{rounded.sm}"
    padding: "7px 12px"
---

# Design System: Dockyard

## 1. Overview

**Creative North Star: "The Instrument Panel"**

Dockyard's visual language is borrowed from precision instruments — cockpit gauges, a ship's control console, a well-maintained server rack with a clear status indicator per unit. The interface is calibrated, not decorated. Nothing is there to impress; everything is there because it earns its position. The whitespace is structural. The typography is exacting. The one accent color is architectural: it holds weight without noise.

Every screen's primary job is to surface state. Are the deployments healthy? Is the build running? Did something fail? That information occupies the visual foreground. Navigation, labels, and structural chrome recede into a quiet cool-white surface. Status signals are the only elements that speak with color — and they speak sparingly, with icon and text alongside the color so the signal is never ambiguous.

The system rejects four failure modes by name: the chart-first density of Grafana, the LED-traffic-light palette of ops dashboards, the startup marketing feel of SaaS cream, and the monospaced-everything of a terminal emulator aesthetic. Dockyard is none of these. It is a professional tool with one user who needs to confirm the state of their infrastructure and act on it — not be impressed, not be entertained, not be hand-held.

**Key Characteristics:**
- Light canvas, cool-white with a barely perceptible blue undertone
- One primary color (warm graphite, near-black) for actions and controls
- Status signals use four semantic colors — subdued, never fluorescent
- Medium density: structured whitespace, not sprawl
- Inter Variable throughout, weight contrast carries hierarchy
- Flat by default — elevation appears only under active layers (dialogs, dropdowns)

## 2. Colors: The Calibrated Palette

A near-monochromatic system where the primary color is dark graphite and the accent work is done entirely by semantic status colors.

### Primary
- **Instrument Graphite** (`oklch(22% 0.03 250)` ≈ `#1c2333`): The primary action color. Used on filled buttons, active navigation indicators, and the logo mark. Carries the authority of near-black without pure coldness. Its near-zero chroma keeps it architectural, not slate-blue.
- **Canvas White** (`oklch(98.5% 0.007 240)` ≈ `#f7f8fa`): The main surface. Not pure white — a whisper of cool blue tint (`chroma 0.007`) that makes text crisp and prevents optical harshness.

### Neutral
- **Deep Foreground** (`oklch(12% 0.025 250)` ≈ `#0e1117`): Body text, headings, data values. Near-black with a cool-blue trace — pairs precisely with the canvas.
- **Cool Slate** (`oklch(48% 0.018 250)` ≈ `#6b7280`): Secondary text, descriptions, placeholder labels. Enough contrast for WCAG AA at this size range.
- **Surface Muted** (`oklch(95.5% 0.008 240)` ≈ `#f1f2f5`): Hover backgrounds, active sidebar items, alternating row fills. One step darker than canvas; the difference is structural.
- **Panel Border** (`oklch(90% 0.010 240)` ≈ `#e3e5ea`): Sidebar rule, card outlines, table dividers, input strokes. Present but quiet.

### Semantic Status Colors
These four colors are not brand colors. They belong to the data, not the chrome.

- **Healthy Green** (`oklch(52% 0.13 145)` ≈ `#2f7a4e`): Deployment `healthy`, build `succeeded`. A forest green — measured, not fluorescent. Used for the status dot and badge background tint only.
- **Building Blue** (`oklch(55% 0.13 230)` ≈ `#3a71b8`): Build `running`, deployment `deploying`. A calibrated instrument blue. Distinct from the graphite primary.
- **Pending Amber** (`oklch(58% 0.12 75)` ≈ `#aa7c1a`): Build `pending`. A warm amber that reads as "waiting" without reading as "warning."
- **Failed Red** (`oklch(52% 0.18 28)` ≈ `#b73b2a`): Deployment `failed`, `rolled_back`, build `failed`. A deep brick red. Serious but not alarming.

### Named Rules
**The One Signal Rule.** Status colors exist for status. Do not use Healthy Green as a brand accent, do not use Building Blue for decorative elements, do not borrow Failed Red for emphasis. Their meaning is fixed to deployment state; diluting that meaning costs comprehension speed.

**The Chroma Ceiling Rule.** No color in this system exceeds `chroma 0.18`. Even the semantic reds and greens are desaturated relative to their hue's maximum. Saturation is spent on legibility, not visual impact.

## 3. Typography: The Instrument Read

**Primary Font:** Inter Variable, Inter, system-ui, sans-serif

A single typeface, varied through weight and size. Inter Variable is the precision instrument of sans-serif typefaces — designed for screens, numerically even, with optional tabular figure settings that make deployment tables and version strings read cleanly.

**Character:** Confident, not decorative. Weight contrast (400 → 600) does the hierarchy work that color restraint cannot. Numbers and technical strings (commit SHAs, port numbers, version tags) deserve tabular figures (`font-variant-numeric: tabular-nums`) wherever they appear in data contexts.

### Hierarchy
- **Display** (600 weight, 1.5rem / 24px, −0.02em tracking): Page-level titles. Used once per view. Examples: "Projects", "Runtime Targets".
- **Headline** (600 weight, 1.125rem / 18px, −0.01em tracking): Section headings within a project detail page, major card titles, dialog headings.
- **Title** (500 weight, 0.9375rem / 15px, normal tracking): Subsection labels, table column headers, tab labels. The workhorse weight-size combination.
- **Body** (400 weight, 0.875rem / 14px, 1.6 line-height): Primary reading text, descriptions, form labels. Max line length 65ch — enforced for description blocks, relaxed for table cells.
- **Label** (500 weight, 0.75rem / 12px, +0.01em tracking): Status labels, badge text, metadata (timestamps, slugs, branch names). Uppercase only for abbreviated category labels (e.g. "SHA", "PORT") — never for running prose.

### Named Rules
**The Tabular Figures Rule.** Any number in a data context — version strings, port numbers, commit SHAs, timestamps in tables — must use `font-variant-numeric: tabular-nums`. Proportional numbers in tables cause column misalignment and look unfinished.

## 4. Elevation

Dockyard is flat by default. Surfaces exist at one tonal level; depth is communicated through border color and background tint, not shadow.

Two elevation steps exist, used only for layered interactive elements:

### Shadow Vocabulary
- **Raised** (`box-shadow: 0 1px 3px rgba(14, 17, 23, 0.08), 0 1px 2px rgba(14, 17, 23, 0.04)`): Cards in hover state, focused dropdown triggers. A barely-there ambient shadow that lifts the element just enough to read as interactive.
- **Overlay** (`box-shadow: 0 8px 32px rgba(14, 17, 23, 0.12), 0 2px 8px rgba(14, 17, 23, 0.06)`): Dialogs, sheet panels, popovers. Communicates modal layering without dramatics.

### Named Rules
**The Flat-By-Default Rule.** Surfaces are flat at rest. Shadows appear only as a response to state (hover, overlay, focus). A resting card has no shadow — only a `1px` border in Panel Border color. Adding shadows to resting elements is prohibited.

## 5. Components

### Buttons

The primary button is the instrument's primary control: dark, decisive, takes up minimal space.

- **Shape:** Gently rounded (5px radius) — not pill, not sharp. The radius is structural, not decorative.
- **Primary:** Instrument Graphite background, near-white text, 7px top/bottom × 14px left/right padding. `font-size: 0.875rem`, `font-weight: 500`.
- **Hover / Focus:** Lightens to `oklch(30% 0.03 250)` on hover (1 step lighter, not a color change). Focus ring: 2px offset ring in Instrument Graphite. `transition: background 150ms ease-out`.
- **Outline:** 1px Panel Border stroke, transparent background, Deep Foreground text. Hover fills with Surface Muted. For secondary actions adjacent to a primary button.
- **Ghost:** No border, transparent background, Cool Slate text. Hover fills with Surface Muted. For tertiary actions (cancel, dismiss).
- **Destructive:** Deep red background (`oklch(52% 0.18 28)`), near-white text. Reserved for irreversible destructive actions. Not a synonym for "important."

### Status Chips (Signature Component)

The most distinctive element in Dockyard. Not a generic badge — a precision reading.

Each status chip shows: a small 6px filled dot (color = semantic status color) + a label in Label weight. The chip background is a 8% tint of the semantic color. No border. The dot + label together carry the signal; neither alone is sufficient.

- **Healthy / Succeeded:** Dot `oklch(52% 0.13 145)`, label "Healthy" or "Succeeded", chip background `oklch(98% 0.03 145)` (barely tinted green).
- **Running / Deploying:** Dot `oklch(55% 0.13 230)`, label "Building" or "Deploying", background `oklch(98% 0.03 230)`. Dot pulses with a slow 2s opacity animation when active.
- **Pending:** Dot `oklch(58% 0.12 75)`, label "Pending", background `oklch(98% 0.04 75)`.
- **Failed / Rolled Back:** Dot `oklch(52% 0.18 28)`, label "Failed", background `oklch(98% 0.05 28)`.

**The Color-Is-Never-Alone Rule.** Every status chip shows dot + text. Color is the quick-scan signal; text is the accessible, unambiguous reading. Chips that show color without text are prohibited.

### Cards / Containers

- **Corner Style:** 8px radius (md) — the structural default.
- **Background:** Canvas White, same level as the page surface.
- **Shadow Strategy:** Flat by default (see Elevation). Hover state only: Raised shadow.
- **Border:** 1px Panel Border (`oklch(90% 0.010 240)`). Always present on resting cards.
- **Internal Padding:** 16px (spacing.md). Header sections may use 20px top if a title + description stack appears.
- **No nested cards.** A card inside a card is always wrong. If a card needs internal grouping, use a subtle divider (`1px Panel Border`) or a Surface Muted background on the grouped region.

### Inputs / Fields

- **Style:** 1px Panel Border stroke, transparent background, 5px radius, 7px × 12px padding.
- **Focus:** Border shifts to Instrument Graphite (`oklch(22% 0.03 250)`), `transition: border-color 120ms ease-out`. No glow, no ring outside the element — the border itself is the focus indicator.
- **Placeholder:** Cool Slate text (`oklch(48% 0.018 250)`).
- **Error state:** Border shifts to Failed Red. An error message appears below in Label size, Failed Red color.
- **Disabled:** 50% opacity on the full field. Cursor: not-allowed.

### Navigation (Sidebar)

- **Width:** 224px (14rem) at rest.
- **Background:** Canvas White with a 1px Panel Border on the right edge.
- **Logo area:** 56px tall, 16px horizontal padding. Anchor icon (⚓) in Instrument Graphite, "Dockyard" in Title weight.
- **Nav items:** Icon (16px) + label in Body weight. Default: Cool Slate text, transparent background. Hover: Surface Muted background, Deep Foreground text. Active: Surface Muted background, Deep Foreground text, Instrument Graphite `3px left inset` border — the only use of a left border accent in this system (inset, not standalone stripe).
- **Footer:** Version string in Label size, Cool Slate. No decoration.

### Tables

- **Row height:** 44px default.
- **Header row:** Label weight, Cool Slate, no background. 1px bottom border in Panel Border.
- **Body rows:** Body weight, Deep Foreground. Alternating rows use Surface Muted background only on hover — not as a static zebra pattern.
- **Dividers:** 1px Panel Border horizontal rules between rows. No vertical dividers.
- **Actions column:** Ghost buttons, right-aligned. Visible only on row hover.

## 6. Do's and Don'ts

### Do:
- **Do** use `font-variant-numeric: tabular-nums` on all numbers in table cells, version strings, port values, and timestamps. Misaligned numbers read as unfinished.
- **Do** show status with dot + text together — never color alone, never text without a color signal.
- **Do** let the border (not a shadow) define a card at rest. Add the Raised shadow on hover only.
- **Do** use Surface Muted (`oklch(95.5% 0.008 240)`) as the hover background for interactive rows, nav items, and ghost buttons — it is the only hover tint in the system.
- **Do** tint the canvas. `oklch(98.5% 0.007 240)` is the background, not `#ffffff`. The difference is subtle; the effect on text sharpness is not.
- **Do** respect `prefers-reduced-motion`. The building/deploying status dot pulse must be suppressed when reduced motion is preferred.

### Don't:
- **Don't** use chart-first layouts, metric dashboards, or data-visualization palette approaches. Dockyard is not Grafana. Status readings are not metrics.
- **Don't** use gradient backgrounds, gradient text (`background-clip: text`), or glass effects. These are SaaS marketing patterns. Dockyard is a precision tool.
- **Don't** use monospaced typefaces as the primary typeface. Mono is appropriate for commit SHAs, CLI output, and code values — nowhere else.
- **Don't** use side-stripe border accents on cards or list items. The one inset-left border in the nav active state is intentional and singular; no other element gets this treatment.
- **Don't** use a rainbow of brand colors. Instrument Graphite is the one brand color. The four semantic status colors belong to the data. If you need a "third" color for anything that isn't a status signal, you are solving the wrong problem.
- **Don't** use fluorescent, high-chroma versions of the status colors. Healthy Green is `oklch(52% 0.13 145)`, not `#00ff00`. Failed Red is `oklch(52% 0.18 28)`, not `#ff0000`. The Chroma Ceiling Rule applies.
- **Don't** stack cards. Nested cards — a card inside a card — are always wrong. Use dividers or a Surface Muted region instead.
- **Don't** use shadows on resting surfaces. The Flat-By-Default Rule is absolute.
- **Don't** add empty-state animations or loading skeleton animations with bouncing/elastic curves. Restrained motion only: opacity fades and linear progress, ease-out-quart at most.
- **Don't** use the Portainer / Docker Desktop visual language: icon-heavy nav, grey sidebar, badge-colored everything, cluttered density.
