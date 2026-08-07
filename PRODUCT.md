# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Existing Next.js + React + TypeScript frontend with a Go API and PostgreSQL persistence.

## Users

- Counselors create and manage client organizations and projects, read assessment information, review stakeholder responses, and help resolve gaps.
- Stakeholders work inside an organization project to answer assigned CSF outcomes and attach supporting evidence.

## Product Purpose

CSF Compliance is a shared NIST CSF 2.0 workspace for organizing an organization's cybersecurity profile, collecting stakeholder input, reviewing responses, and tracking current versus target coverage. Success means a counselor can understand the state of a project quickly while a stakeholder can complete the requested inputs without confusion.

## Positioning

The product separates counselor-owned assessment/profile decisions from stakeholder-owned responses and evidence while keeping both views in the same project workspace.

## Operating Context

- A Counselor Admin creates organizations and projects.
- Counselors work across client projects and need a reading-oriented overview of assessment progress.
- Stakeholders open a project, see the relevant outcomes, enter responses, upload evidence, and submit for review.
- Reviewers can review submitted responses; viewers can read without mutation controls.
- The product is used for long sessions with substantial compliance text and form content.

## Capabilities and Constraints

- Email/password authentication and role-based access are already implemented.
- NIST CSF 2.0 outcomes are organized by Function and assessment cards.
- Current Profile and Target Profile information must remain distinct.
- Stakeholders must not gain counselor-only editing controls through the visual redesign.
- The redesign must preserve existing API, workflow, role permissions, calculations, and content terminology.

## Brand Commitments

- The interface should follow a Clean Editorial Layout and Reader-Friendly Layout direction.
- Use the 60/30/10 color rule with white as the dominant reading surface, light neutral structure, and a restrained teal accent.
- Do not add visual noise that competes with long-form compliance reading.

## Evidence on Hand

- Existing frontend source under `web/src`.
- NIST CSF 2.0 compliance workflow and ER diagrams under the project documentation/download references.
- Existing role and stakeholder response tests that define behavior which the redesign must preserve.

## Product Principles

- Make the next reading or input action obvious.
- Keep counselor interpretation and stakeholder input visibly separate.
- Prefer clear grouping and calm rhythm over decorative chrome.
- Make status and permissions legible without relying on color alone.

## Accessibility & Inclusion

- The interface must support long reading sessions with comfortable type, line-height, spacing, and contrast.
- Keyboard focus, responsive reflow, and readable empty/error states are required.
