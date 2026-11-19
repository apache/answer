---
version: alpha
name: Apache Answer
description: A React Bootstrap interface for Q&A communities, branded through the primary color.
omitted:
  - section: typography
    reason: Inherits Bootstrap typography without separate brand typography tokens.
  - section: spacing
    reason: Uses Bootstrap spacing utilities and existing page layouts without a separate spacing scale.
  - section: rounded
    reason: Inherits default Bootstrap component radii.
  - section: motion
    reason: Inherits Bootstrap and React Bootstrap component transitions.
---

# Apache Answer Design System

## Overview

Apache Answer supports Q&A communities and knowledge sharing. Keep the interface clear and practical, prioritizing readable questions, answers, and actions.

The foundation is **React Bootstrap 2.10.6 + Bootstrap 5.3.3** (currently locked versions). Brand customization changes only the primary color. Component structure, sizing, typography, radii, and interactions follow Bootstrap defaults. Reuse existing domain styles and layouts without expanding them into a separate component design system.

Dependency sources: `@ui/package.json` and `@ui/pnpm-lock.yaml`. Global stylesheet entry point: `@ui/src/index.scss`.

## Colors

The default brand primary is **Answer Blue / `#0033FF`**, used for primary actions, links, selected states, and branded navigation. Success, warning, danger, secondary actions, and neutrals retain Bootstrap semantic colors. Do not turn every state blue.

```yaml
colors:
  primary: "#0033FF"
```

This token represents the default theme. The site setting `theme_config[theme].primary_color` can override the active primary color. Design previews should reflect the target site's configuration.

At build time, `@ui/src/common/_variable.scss` sets the default primary through `$blue` before Bootstrap is imported. At runtime, `@ui/src/components/CustomizeTheme/index.tsx` synchronizes primary colors, links, buttons, pagination, selected items, and form focus states. Reuse this mechanism to derive hover, active, focus, and disabled states instead of introducing isolated color values.

Light and dark modes use `data-bs-theme`. Bootstrap semantic variables and `@ui/src/common/color.scss` adapt surfaces, text, and borders. Dark mode links use lighter primary shades derived by the existing theme logic. Avoid hardcoded light backgrounds in components.

## Typography

Use Bootstrap's default system font stack, sizes, weights, and line heights. Establish hierarchy with semantic headings, body text, and existing utilities such as `small`. Keep the existing typography scale and system fonts. Project utilities such as `.small` and `.fs-12` follow their global stylesheet definitions.

Long questions, answers, and code should wrap or scroll locally without overflowing the page. Preserve semantic headings, form labels, and body structure.

## Layout

Use Bootstrap responsive containers, grids, breakpoints, and spacing utilities. Reuse neighboring page layouts first. Pages typically combine navigation, main content, and an optional sidebar. Stack content naturally on narrow screens and keep primary actions accessible.

The existing sidebar layout is defined in `@ui/src/common/sideNavLayout.scss`: `.main-mx-with` has a maximum width of `1072px`; `.page-right-side` has a desktop width of `300px` and becomes full width below `1199.9px`. Inherit these values when using that layout, rather than imposing them on every page.

## Elevation & Depth

Use Bootstrap borders, surface layers, and default overlay styles to distinguish regions. Preserve the existing navigation shadow. Keep ordinary content areas free of additional decorative shadows. Dropdowns, tooltips, and modals follow component defaults for stacking and positioning.

## Shapes

Buttons, inputs, cards, and modals retain Bootstrap default radii and borders. Domain elements such as avatars and tags reuse their existing component shapes. Do not apply a universal pill or square treatment.

## Motion & Micro-interaction

Use React Bootstrap's built-in collapse, fade, and component state feedback. Preserve default transition timing and honor `prefers-reduced-motion`. Avoid decorative looping animation. Asynchronous actions should communicate loading and prevent duplicate submission.

## Voice & Content Strategy

Write direct, friendly copy with clear action names and consistent Q&A terminology. Reuse existing translation keys in `i18n/`. Add new copy through the internationalization workflow rather than hardcoding a single language in components.

Action labels describe outcomes, such as “Post question” and “Save changes.” Empty states explain what is missing and suggest a next step. Errors explain the failure and an actionable remedy. Dates, numbers, and capitalization follow the active locale and neighboring interfaces.

## Components

Prefer components exported by `react-bootstrap` and reuse existing domain wrappers in `ui/src/components/`.

- **Buttons:** Use `Button variant="primary"` for primary actions. Use existing secondary, outline, or link variants for secondary actions. Preserve danger semantics for destructive actions.
- **Forms:** Use standard `Form`, `Form.Control`, and `Form.Select` components. Keep labels, help text, validation messages, and visible focus. Checked states and focus colors follow the theme.
- **Navigation and pagination:** Reuse Nav, Navbar, Dropdown, and Pagination. Selected states follow primary; sizing and keyboard behavior follow component defaults.
- **Content and feedback:** Reuse Card, ListGroup, Alert, Modal, and Tooltip. Question tags retain existing tag components and `.badge-tag` styling.
- **Icons:** Use the project's existing Bootstrap Icons. Give icon buttons accessible names.

These tokens describe only the brand color mapping for default primary buttons. Bootstrap supplies other properties, and the existing theme implementation derives interaction states.

```yaml
components:
  button-primary:
    backgroundColor: "{colors.primary}"
  button-outline-primary:
    textColor: "{colors.primary}"
```

## Do's and Don'ts

- Prefer component props, Bootstrap utilities, and semantic CSS variables.
- Maintain primary at the theme entry points; avoid repeating hardcoded `#0033FF` in domain components.
- Preserve hover, focus, active, disabled, loading, and validation states.
- Check light mode, dark mode, narrow screens, and keyboard interaction. Keep text, icons, and focus indicators clearly visible.
- Keep the existing font, radius, shadow, spacing, and button systems.
- Preserve danger, success, and warning semantics. Communicate status through more than color alone.
