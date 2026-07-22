# GK Circle Frontend Rules

Version: 1.0

Status: Mandatory

---

# Purpose

These rules govern all frontend development in GK Circle.

All pages, components, layouts, dashboards, landing pages, admin panels, mobile interfaces, and AI-generated UI must follow these standards.

Consistency is more important than individual preferences.

---

# Frontend Technology Stack

Approved Stack:

* Next.js App Router
* React
* TypeScript
* TailwindCSS
* ShadCN UI
* Framer Motion
* React Hook Form
* Zod
* TanStack Query
* Zustand (only when required)

Do not introduce alternative frameworks without approval.

---

# Mobile First Rule

All UI must be designed mobile-first.

Design order:

Mobile

↓

Tablet

↓

Desktop

Desktop-first design is prohibited.

---

# Design System Rule

All UI must use:

docs/design-system.md

as the source of truth.

Do not create independent styling systems.

Do not create page-specific themes.

Do not create isolated design languages.

---

# Global Theme Rule

Styling must originate from:

frontend/src/theme/

or approved design-system files.

Components must consume theme tokens.

Never hardcode colors, spacing, shadows, or typography repeatedly.

---

# Reusable Component Rule

Before creating a component:

Search:

frontend/components/

If component exists:

Reuse it.

If similar component exists:

Extend it.

Do not duplicate components.

---

# Component Hierarchy Rule

Structure:

ui/

↓

shared/

↓

features/

↓

pages/

Reusable components belong higher.

Feature-specific components belong lower.

---

# Page Composition Rule

Pages should compose components.

Pages should NOT contain:

* Business logic
* Large forms
* API calls
* Complex state management

Move logic into:

hooks/

services/

features/

---

# Single Responsibility Rule

Each component should do one thing well.

Avoid:

Mega Components

1000+ line components

Mixed concerns

---

# Animation Rule

Animations must be:

Meaningful

Intentional

Performance-safe

Accessible

Avoid decorative animations that do not improve UX.

---

# Approved Animation Stack

Use:

Framer Motion

Avoid:

Custom animation libraries

Heavy animation frameworks

unless approved.

---

# Landing Page Rule

Landing pages must:

* Follow GK Circle brand identity
* Be premium quality
* Be mobile-first
* Use reusable components
* Use design tokens
* Use Framer Motion
* Use performance-optimized animations

Never generate generic SaaS templates.

---

# Dashboard Rule

Dashboards must:

* Feel premium
* Follow design system
* Use reusable cards
* Support responsive layouts
* Avoid generic admin-template appearance

Dashboard design should reflect a modern AI-powered educational platform.

---

# Accessibility Rule

Every UI must support:

Keyboard Navigation

Visible Focus States

Screen Readers

Semantic HTML

Color Contrast Standards

ARIA attributes where necessary

Accessibility is mandatory.

---

# Typography Rule

Typography must come from:

Design System

Do not create custom typography hierarchies per page.

Maintain consistency.

---

# Color Rule

Colors must originate from:

Theme Tokens

Do not hardcode:

hex values

rgb values

random colors

inside components.

---

# Dark Mode Rule

All major interfaces should support:

Light Mode

Dark Mode

through theme tokens.

Do not create separate component versions.

---

# Form Rule

Forms must use:

React Hook Form

*

Zod Validation

Validation must exist:

Frontend

AND

Backend

---

# API Integration Rule

Components must not call APIs directly.

Use:

services/

hooks/

TanStack Query

for data access.

---

# State Management Rule

Use:

React State

↓

Context

↓

Zustand

Only escalate when necessary.

Avoid global state when local state is sufficient.

---

# Error Handling Rule

Every async UI must handle:

Loading

Success

Empty State

Error State

Do not assume success.

---

# Empty State Rule

Every list page must handle:

No Data

No Search Results

No Permissions

Offline

Gracefully.

---

# Skeleton Loading Rule

Use:

Skeleton Components

instead of loading spinners when appropriate.

Perceived performance matters.

---

# Responsive Layout Rule

Support:

320px

375px

768px

1024px

1440px+

Avoid layouts that only work on desktop.

---

# Performance Rule

Avoid:

Unnecessary rerenders

Large bundles

Deep prop drilling

Heavy animations

Unused dependencies

Optimize:

Images

Fonts

Queries

Components

---

# Image Rule

Use:

next/image

for all images.

Optimize loading.

Avoid unoptimized image tags.

---

# Feature Folder Rule

Feature-specific code belongs in:

features/

Examples:

features/Courses

features/auth

features/tests

features/communities

Do not place feature logic inside pages.

---

# Design Consistency Rule

New pages must visually match:

Existing approved pages.

Do not create isolated UI styles.

Do not create visual fragmentation.

---

# Playwright Readiness Rule

Every UI must be testable.

Provide:

Stable selectors

Accessible labels

Predictable states

Avoid fragile UI structures.

---

# AI Generated UI Rule

AI-generated UI must:

Use existing design system

Use existing components

Use theme tokens

Use responsive layouts

Use accessibility standards

AI must not invent a separate design language.

---

# Completion Rule

Frontend work is complete only when:

✓ Responsive

✓ Accessible

✓ Tested

✓ Integrated

✓ Themed

✓ Verified

✓ Screenshots captured

✓ Playwright passed

---

# Final Directive

Build interfaces users love.

Build systems developers can maintain.

Consistency beats creativity.

Quality beats speed.

User experience beats visual noise.
