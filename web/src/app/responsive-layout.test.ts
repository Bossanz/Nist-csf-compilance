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
  expect(css).toContain(".project-context-overview { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr));");
  expect(css).toContain(".project-metadata { display: grid;");
  expect(css).toMatch(/\.project-metadata > div \{[^}]*align-content: start;[^}]*align-items: start;/);
  expect(css).toContain(".cards { grid-template-columns: repeat(4, minmax(0, 1fr)); }");
  expect(css).toContain(".coverage-value { display: grid;");
  expect(css).toContain(".evidence-count {");
});

test("centers the project workspace and uses available desktop width", () => {
  expect(css).toContain("--content-max: 1320px;");
  expect(css).toContain(".project-header { width: min(100%, var(--content-max)); margin: 0 auto 30px; }");
  expect(css).toMatch(/\.project-layout \{[^}]*width: min\(100%, var\(--content-max\)\);[^}]*margin: 0 auto;/);
  expect(css).toContain(".reading-column { width: 100%; }");
});

test("shares a readable content measure and page gutter", () => {
  expect(css).toContain("--page-gutter: clamp(22px, 4vw, 64px);");
  expect(css).toContain("--reading-measure: 72ch;");
  expect(css).toContain("padding: 40px var(--page-gutter) 80px;");
  expect(css).toContain(".dashboard { width: min(100%, var(--content-max)); max-width: var(--content-max); margin: 0 auto; }");
  expect(css).toContain("p { max-width: var(--reading-measure); }");
});

test("uses semantic tokens for contextual colors and focus", () => {
  expect(css).toContain("--focus-outline: rgba(230, 0, 122, 0.72);");
  expect(css).toContain("--accent-line: #f0a7cd;");
  expect(css).toContain("--accent-line-soft: #f5c5df;");
  expect(css).toContain("--accent-line-faint: #f9ddec;");
  expect(css).toContain("--included-line: #dda8c8;");
  expect(css).toContain("--input-hover: #b86b9b;");
  expect(css).toContain("--error-line: #e8c3bf;");
  expect(css).toContain("--current-wash: rgba(237, 242, 255, 0.7);");
  expect(css).toContain("--target-wash: rgba(255, 243, 220, 0.72);");
  expect(css).toContain("outline: 3px solid var(--focus-outline);");
  expect(css).toContain("border-color: var(--accent-line);");
  expect(css).toContain("border-color: var(--input-hover);");
  expect(css).toContain("border: 1px solid var(--error-line);");
  expect(css).toContain("background: var(--current-wash);");
  expect(css).toContain("background: var(--target-wash);");
});

test("stacks project context and evidence actions at narrow widths", () => {
  expect(css).toContain(".project-context-overview { grid-template-columns: 1fr; gap: 12px; }");
  expect(css).toContain(".project-metadata { grid-template-columns: 1fr; gap: 10px; }");
  expect(css).toContain(".cards, .summary-strip { grid-template-columns: repeat(2, minmax(0, 1fr)); }");
  expect(css).toContain(".evidence-list li .button-link { max-width: 100%; }");
  expect(css).toContain(".response-buttons button { flex: 1; }");
  expect(css).toContain("@media (prefers-reduced-motion: reduce)");
});
test("keeps mobile assessment content readable and stacked", () => {
  expect(css).toContain(".project-card { min-height: auto; gap: 18px; padding: 16px; }");
  expect(css).toContain(".project-card-top { align-items: flex-start; flex-direction: column; gap: 4px; }");
  expect(css).toMatch(/\.assessment-summary \{[^}]*grid-template-columns: 88px minmax\(0, 1fr\) 30px;/);
  expect(css).toMatch(/\.coverage-route \{[^}]*grid-column: 1 \/ -1;[^}]*flex-wrap: wrap;/);
});

test("uses a targeted reduced-motion alternative", () => {
  expect(css).toContain(".assessment-body { animation: none; }");
  expect(css).toContain("button.primary, button.secondary, .nav-item, input, select, textarea, .anchor-primary { transition: none; }");
  expect(css).not.toContain("*, *::before, *::after { animation-duration: 0.01ms !important; transition-duration: 0.01ms !important; }");
});

test("uses restrained accent rules instead of thick side tabs", () => {
  const thickSideTab = ["border-left", "3px solid var(--accent);"].join(": ");
  expect(css).toContain("border-left: 1px solid var(--accent);");
  expect(css).not.toContain(thickSideTab);
});
