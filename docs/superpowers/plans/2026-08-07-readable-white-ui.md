# Readable White UI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ปรับหน้าจอหลักของ CSF Compliance ให้เป็น white-first Calm Editorial UI ที่อ่านข้อมูล compliance จำนวนมากได้สบายขึ้น โดยไม่เปลี่ยน API, role หรือ business logic

**Architecture:** คง Next.js App Router และ client component composition เดิมไว้ ใช้ CSS tokens ใน `globals.css` เป็น source of truth แล้วปรับ markup เฉพาะจุดที่ช่วยลำดับข้อมูลและ accessibility. Assessment ยังคงเป็น accordion เดิม แต่ทำให้ summary, detail, Current/Target และ response panel มี hierarchy ที่ชัดขึ้น

**Tech Stack:** Next.js 16, React 19, TypeScript, plain CSS, Vitest, Testing Library

## Global Constraints

- 60% สีขาวสำหรับพื้นที่อ่าน การ์ด ฟอร์ม และรายละเอียดของ assessment
- 30% สีเทาอ่อน/เขียวเทาอ่อนสำหรับ canvas, section ย่อย และพื้นที่พักสายตา
- 10% สี teal สำหรับ primary action, เมนูที่เลือก, สถานะสำคัญ และเส้นนำสายตา
- แก้ใน global tokens/styles และ component markup เท่าที่จำเป็น
- ไม่เพิ่ม UI kit, icon set หรือ animation package
- ไม่เปลี่ยน schema, API, role permission, workflow หรือ business calculation
- รองรับ `prefers-reduced-motion` และ focus-visible ที่มองเห็นชัด

---

### Task 1: Establish white-first tokens and app/auth shell

**Files:**
- Modify: `web/src/app/globals.css`
- Modify: `web/src/components/FunctionSidebar.tsx`
- Modify: `web/src/components/LoginForm.tsx` only if a semantic wrapper needs a class adjustment
- Test: `web/src/components/FunctionSidebar.test.tsx`

**Interfaces:**
- Consumes: existing `.shell`, `.sidebar`, `.main`, `.auth-layout`, `.auth-panel`, `.nav-item`, `.field`, `.primary`, `.secondary` selectors
- Produces: centralized white-first tokens and a labeled function navigation landmark; all existing component behavior remains unchanged

- [ ] **Step 1: Write the failing accessibility test**

Create `web/src/components/FunctionSidebar.test.tsx` with a focused assertion for the navigation landmark:

```tsx
import { render, screen } from "@testing-library/react";
import { expect, test } from "vitest";
import { FunctionSidebar } from "./FunctionSidebar";

test("labels the function navigation landmark", () => {
  render(<FunctionSidebar functions={[{ id: "fn-1", code: "GV", name: "Govern", description: "", categories: [] }]} selectedCode="GV" onSelect={() => undefined} />);

  expect(screen.getByRole("navigation", { name: /csf functions/i })).toBeTruthy();
});
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run from `web`:

```powershell
npm.cmd test -- --run src/components/FunctionSidebar.test.tsx
```

Expected: FAIL because the existing sidebar is an unlabeled `<aside>` and is not exposed as a named navigation landmark.

- [ ] **Step 3: Implement the minimal shell/accessibility change**

Change the function sidebar wrapper to:

```tsx
<nav className="sidebar" aria-label="CSF functions">
```

Then update `globals.css` tokens and base rules so that `--canvas` is a very light neutral, `--surface` remains white, borders are low-contrast, body text uses a readable system stack and `line-height: 1.6`. Keep teal as the only accent and preserve existing semantic error/success colors.

- [ ] **Step 4: Run the focused test and existing auth tests**

```powershell
npm.cmd test -- --run src/components/FunctionSidebar.test.tsx src/components/LoginForm.test.tsx
```

Expected: PASS with no behavior changes to login or function selection.

- [ ] **Step 5: Commit the shell slice**

```powershell
git add web/src/app/globals.css web/src/components/FunctionSidebar.tsx web/src/components/FunctionSidebar.test.tsx
git commit -m "style: establish readable white app shell"
```

### Task 2: Clarify organization, workspace and project hierarchy

**Files:**
- Modify: `web/src/components/OrganizationDashboard.tsx`
- Modify: `web/src/components/OrganizationWorkspace.tsx`
- Modify: `web/src/components/ProjectDashboard.tsx`
- Modify: `web/src/components/SummaryCards.tsx`
- Modify: `web/src/app/globals.css`
- Test: existing `web/src/components/OrganizationDashboard.test.tsx`, `web/src/components/OrganizationWorkspace.test.tsx`, and `web/src/components/ProjectDashboard.test.tsx`

**Interfaces:**
- Consumes: existing callbacks and `Organization`, `Project`, `User`, and `Summary` props
- Produces: consistent page headers, readable list/card surfaces, one primary action per create section, and clearer secondary/danger actions without changing callback signatures

- [ ] **Step 1: Add structure assertions to existing component tests**

Extend the organization and project tests with assertions that the existing page headings remain accessible and that the create sections expose their named headings:

```tsx
expect(screen.getByRole("heading", { name: /organizations/i })).toBeTruthy();
expect(screen.getByRole("heading", { name: /create organization/i })).toBeTruthy();
```

Keep the current interaction assertions unchanged; these tests protect the hierarchy while the markup is cleaned up.

- [ ] **Step 2: Run the affected tests before styling**

```powershell
npm.cmd test -- --run src/components/OrganizationDashboard.test.tsx src/components/OrganizationWorkspace.test.tsx src/components/ProjectDashboard.test.tsx
```

Expected: PASS before the visual changes.

- [ ] **Step 3: Refine the page markup without changing behavior**

Keep `main`, `header`, `section`, `h1`, and `h2` landmarks. Add or retain stable classes for:

```tsx
<header className="product-header">...</header>
<section className="content-section" aria-labelledby="...">...</section>
<section className="panel create-organization">...</section>
```

Use the same pattern for workspace/project pages. Keep destructive actions as `.danger`, create/open actions as `.primary` or `.secondary`, and do not add a second confirmation flow.

- [ ] **Step 4: Implement the hierarchy styles**

In `globals.css`:

- make the main content column wider but capped for reading;
- give page headers consistent bottom spacing and a quiet identity block;
- use white cards with a 1px border and minimal shadow;
- make grids collapse to one column earlier on narrow screens;
- make summary cards lighter, more scannable and less visually dominant than the page heading;
- style project/organization numbers and teal rails as navigation cues, not decoration;
- preserve readable loading, empty and error states.

- [ ] **Step 5: Run the affected tests and TypeScript check**

```powershell
npm.cmd test -- --run src/components/OrganizationDashboard.test.tsx src/components/OrganizationWorkspace.test.tsx src/components/ProjectDashboard.test.tsx
npm.cmd run typecheck
```

Expected: PASS with unchanged callback behavior.

- [ ] **Step 6: Commit the workspace slice**

```powershell
git add web/src/components/OrganizationDashboard.tsx web/src/components/OrganizationWorkspace.tsx web/src/components/ProjectDashboard.tsx web/src/components/SummaryCards.tsx web/src/app/globals.css
git commit -m "style: organize workspace and project views"
```

### Task 3: Improve assessment reading flow and response hierarchy

**Files:**
- Modify: `web/src/components/AssessmentCard.tsx`
- Modify: `web/src/components/StakeholderResponsePanel.tsx`
- Modify: `web/src/components/ProfileEditor.tsx` only if the list needs a semantic label
- Modify: `web/src/app/globals.css`
- Test: existing `web/src/components/ProfileEditor.test.tsx` and `web/src/components/StakeholderResponsePanel.test.tsx`

**Interfaces:**
- Consumes: existing `ProfileRow`, `ProfilePatch`, `StakeholderResponse`, `ResponseDocument`, and role-based callbacks
- Produces: the same save/review/upload behavior with improved visual hierarchy and clear read-only states

- [ ] **Step 1: Add a focused accessibility assertion for the assessment list**

Extend `ProfileEditor.test.tsx` with:

```tsx
expect(screen.getByRole("button", { name: /GV\.OC-01/i })).toHaveAttribute("aria-expanded", "false");
```

This codifies the approved collapsed-by-default reading flow before styling changes.

- [ ] **Step 2: Run the focused assessment tests**

```powershell
npm.cmd test -- --run src/components/ProfileEditor.test.tsx src/components/StakeholderResponsePanel.test.tsx
```

Expected: PASS; the existing accordion and role controls remain intact.

- [ ] **Step 3: Preserve the accordion and sharpen its summary hierarchy**

Keep the current summary button and controlled `expanded` state. Use the existing `assessment-summary` classes to give the outcome code, description, coverage route and expand mark distinct levels. Keep included items on the teal reading rail and excluded items on a quiet neutral rail.

- [ ] **Step 4: Style the detail and response surfaces for long reading**

Update `globals.css` so that:

- the expanded detail has comfortable padding and line-height;
- Current and Target panels use very light tinted surfaces with clear headings;
- fields have generous label-to-control spacing and visible focus rings;
- the response panel is visually subordinate but still easy to find;
- reviewer controls, evidence list, save state and error state remain readable on white;
- mobile changes Current/Target and response actions to a single column without clipping.

- [ ] **Step 5: Run assessment tests and TypeScript check**

```powershell
npm.cmd test -- --run src/components/ProfileEditor.test.tsx src/components/StakeholderResponsePanel.test.tsx
npm.cmd run typecheck
```

Expected: PASS with unchanged permissions and callbacks.

- [ ] **Step 6: Commit the assessment slice**

```powershell
git add web/src/components/AssessmentCard.tsx web/src/components/StakeholderResponsePanel.tsx web/src/components/ProfileEditor.tsx web/src/app/globals.css web/src/components/ProfileEditor.test.tsx
git commit -m "style: make assessment reading flow clearer"
```

### Task 4: Full frontend verification and responsive review

**Files:**
- Verify: `web/src/app/globals.css`
- Verify: all changed frontend components and tests from Tasks 1-3

**Interfaces:**
- Consumes: completed UI slices
- Produces: evidence that the white-first redesign works at desktop/mobile breakpoints and preserves the existing frontend contract

- [ ] **Step 1: Run the complete frontend test suite**

```powershell
npm.cmd test -- --run
npm.cmd run typecheck
```

Expected: all Vitest tests pass and TypeScript reports no errors.

- [ ] **Step 2: Check formatting and accidental changes**

```powershell
git diff --check
git status --short
```

Expected: no whitespace errors, no generated dependency files, and only the intended frontend/spec/plan files changed.

- [ ] **Step 3: Review the running app at desktop and mobile widths**

Run the existing project with `docker compose up -d` if it is not already running, then inspect `/` and exercise: login, organization list, workspace, project list, and one expanded assessment. Confirm that cards do not collide, long descriptions wrap, buttons remain reachable, and the collapsed assessment list is easy to scan.

- [ ] **Step 4: Commit the verified final UI**

```powershell
git add web/src/app/globals.css web/src/components/FunctionSidebar.tsx web/src/components/FunctionSidebar.test.tsx web/src/components/LoginForm.tsx web/src/components/OrganizationDashboard.tsx web/src/components/OrganizationWorkspace.tsx web/src/components/ProjectDashboard.tsx web/src/components/SummaryCards.tsx web/src/components/AssessmentCard.tsx web/src/components/StakeholderResponsePanel.tsx web/src/components/ProfileEditor.tsx web/src/components/ProfileEditor.test.tsx
git commit -m "style: polish readable compliance workspace"
```
