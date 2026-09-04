import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { expect, test } from "vitest";

const css = readFileSync(resolve(process.cwd(), "src/app/globals.css"), "utf8");

test("adapts organization controls for tablet widths", () => {
  expect(css).toContain(".invite-form { grid-template-columns: minmax(0, 1fr) minmax(180px, 0.8fr); }");
  expect(css).toContain(".invite-form .primary { grid-column: 1 / -1; justify-self: start; }");
});

test("lets repeated workspaces use available width without forcing empty columns", () => {
  expect(css).toMatch(/\.organization-grid, \.project-grid \{[^}]*grid-template-columns: repeat\(auto-fit, minmax\(min\(100%, 420px\), 1fr\)\);/);
});

test("gives a single workspace card a focused reading width", () => {
  expect(css).toContain(".organization-grid:has(> :only-child), .project-grid:has(> :only-child) { grid-template-columns: minmax(0, 760px); }");
});

test("gives mobile organization and project actions full-width touch targets", () => {
  expect(css).toContain(".organization-actions, .project-actions { width: 100%; justify-content: stretch; }");
  expect(css).toContain(".organization-actions button, .project-actions button { flex: 1; }");
  expect(css).toContain(".organization-card { grid-column: 1 / -1; }");
});

test("gives the role-first project context and outcome summary readable structure", () => {
  expect(css).toContain(".project-context-panel {");
  expect(css).not.toContain(".project-context-overview");
  expect(css).toContain(".project-metadata { display: grid;");
  expect(css).toMatch(/\.project-metadata > div \{[^}]*align-content: start;[^}]*align-items: start;/);
  expect(css).toContain(".cards { grid-template-columns: repeat(4, minmax(0, 1fr)); }");
  expect(css).toContain(".coverage-value { display: grid;");
  expect(css).toContain(".evidence-count {");
});

test("gives project finalization a clear responsive layout", () => {
  expect(css).toContain(".project-finalization-panel {");
  expect(css).toContain(".finalization-readiness {");
  expect(css).toContain(".finalization-blocker {");
  expect(css).toContain(".finalized-panel {");
});

test("centers the project workspace and uses available desktop width", () => {
  expect(css).toContain("--qid-content-max: 1200px;");
  expect(css).toContain(".project-header { width: min(100%, var(--content-max)); margin: 0 auto 30px; }");
  expect(css).toMatch(/\.project-layout \{[^}]*width: min\(100%, var\(--content-max\)\);[^}]*margin: 0 auto;/);
  expect(css).toContain(".reading-column { width: 100%; }");
});

test("shares a readable content measure and page gutter", () => {
  expect(css).toContain("--qid-page-gutter: clamp(20px, 4vw, 48px);");
  expect(css).toContain("--qid-reading-measure: 68ch;");
  expect(css).toContain("padding: 40px var(--qid-page-gutter) 80px;");
  expect(css).toContain(".dashboard { width: min(100%, var(--content-max)); max-width: var(--content-max); margin: 0 auto; }");
  expect(css).toContain("p { max-width: var(--reading-measure); }");
});

test("uses semantic tokens for contextual colors and focus", () => {
  expect(css).toContain("--qid-focus-ring: 0 0 0 3px color-mix(in srgb, var(--qid-pink) 24%, transparent);");
  expect(css).toContain("--qid-current-text: #a9c7ff;");
  expect(css).toContain("--qid-target-text: #ffd58a;");
  expect(css).toContain("--qid-border-accent: #eb147c;");
  expect(css).toContain("--qid-border-soft: #f3b1d2;");
  expect(css).toContain("--qid-border-faint: #f8d8e9;");
  expect(css).toContain("--qid-current-wash: color-mix(in srgb, var(--qid-current) 10%, var(--qid-surface));");
  expect(css).toContain("--qid-target-wash: color-mix(in srgb, var(--qid-target) 13%, var(--qid-surface));");
  expect(css).toContain("outline: 3px solid var(--focus-outline);");
  expect(css).toContain("border-color: var(--accent-line);");
  expect(css).toContain("border-color: var(--input-hover);");
  expect(css).toContain("border: 1px solid var(--error-line);");
  expect(css).toContain("background: var(--current-wash);");
  expect(css).toContain("background: var(--target-wash);");
  expect(css).toContain(".current-coverage strong { color: var(--qid-current-text); }");
  expect(css).toContain(".target-coverage strong { color: var(--qid-target-text); }");
  expect(css).toContain(".status-submitted { color: var(--qid-current-text);");
});

test("stacks project context and evidence actions at narrow widths", () => {
  expect(css).toContain(".project-metadata { grid-template-columns: 1fr; gap: 10px; }");
  expect(css).toContain(".cards, .summary-strip { grid-template-columns: repeat(2, minmax(0, 1fr)); }");
  expect(css).toContain(".evidence-list li .button-link { max-width: 100%; overflow: visible; overflow-wrap: anywhere; text-align: left; white-space: normal; }");
  expect(css).toContain(".response-buttons button { flex: 1; }");
  expect(css).toContain("@media (prefers-reduced-motion: reduce)");
});
test("keeps mobile assessment content readable and stacked", () => {
  expect(css).toContain(".project-card { min-height: auto; gap: 18px; padding: 16px; }");
  expect(css).toContain(".project-card-top { align-items: flex-start; flex-direction: column; gap: 4px; }");
  expect(css).toMatch(/\.assessment-summary \{[^}]*grid-template-columns: 88px minmax\(0, 1fr\) 30px;/);
  expect(css).toMatch(/\.coverage-route \{[^}]*grid-column: 1 \/ -1;[^}]*flex-wrap: wrap;/);
});

test("lets long account and evidence names wrap on mobile", () => {
  expect(css).toContain(".person-row span:not(.role-chip) { overflow: visible; overflow-wrap: anywhere; text-overflow: clip; white-space: normal; }");
  expect(css).toContain(".invitation-meta { justify-items: start; white-space: normal; }");
  expect(css).toContain(".evidence-list li .button-link { max-width: 100%; overflow: visible; overflow-wrap: anywhere; text-align: left; white-space: normal; }");
});

test("uses a targeted reduced-motion alternative", () => {
  expect(css).toContain(".assessment-body { animation: none; }");
  expect(css).toContain("button.primary, button.secondary, .sidebar-link, .sidebar-function, .sidebar-icon-button, input, select, textarea, .anchor-primary { transition: none; }");
  expect(css).not.toContain("*, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; }");
});

test("keeps invitation and workspace navigation controls at the shared touch height", () => {
  expect(css).toContain(".invitation-actions button { min-height: 44px;");
  expect(css).toMatch(/\.sidebar-link \{[\s\S]*min-height: 44px;/);
  expect(css).toMatch(/\.sidebar-icon-button \{[\s\S]*min-height: 44px;/);
});

test("stacks the Evidence Workbench registers at narrow widths", () => {
  expect(css).toContain(".organization-register-row, .workspace-register-row { grid-template-columns: 1fr;");
  expect(css).toContain(".function-register-row { grid-template-columns: 1fr;");
  expect(css).toContain(".next-action-panel { grid-template-columns: 1fr;");
});

test("keeps workbench surfaces on the QID token system", () => {
  expect(css).toContain(".overview-workbench {");
  expect(css).toContain("background: var(--qid-surface);");
  expect(css).toContain("border-top: 2px solid var(--qid-pink);");
  expect(css).toContain(".function-register-progress-track span { display: block; height: 100%; background: var(--qid-action-gradient);");
});

test("makes dense report tables discoverable and keyboard-scrollable on small screens", () => {
  expect(css).toContain(".report-table-wrap { position: relative;");
  expect(css).toContain(".report-table-wrap:focus-visible");
  expect(css).toContain(".report-table-hint");
  expect(css).toContain(".report-table-wrap::after");
});

test("uses restrained accent rules instead of thick side tabs", () => {
  const thickSideTab = ["border-left", "3px solid var(--accent);"].join(": ");
  expect(css).toContain("border-left: 1px solid var(--accent);");
  expect(css).not.toContain(thickSideTab);
});
