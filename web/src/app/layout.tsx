import "./globals.css";
import type { Metadata } from "next";
export const metadata: Metadata = { title: "CSF Compliance", description: "NIST CSF 2.0 compliance workspace" };

const designContract = "Evidence Workbench | QID v3 Compliance Workspace | dark-first workspace system | status-first registers | magenta-purple action gradient";
const themeBootstrapScript = `
(() => {
  try {
    const theme = window.localStorage.getItem("csf-theme");
    document.documentElement.dataset.theme = theme === "light" ? "light" : "dark";
    document.documentElement.style.colorScheme = theme === "light" ? "light" : "dark";
  } catch {
    document.documentElement.dataset.theme = "dark";
    document.documentElement.style.colorScheme = "dark";
  }
})();
`;

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" data-theme="dark" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeBootstrapScript }} />
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&family=Space+Grotesk:wght@400;500;600;700&display=swap"
          rel="stylesheet"
        />
      </head>
      <body data-design-contract={designContract} data-design-direction="evidence-workbench-250b41bb">
        {/*
        THESIS: Turn a complex compliance assessment into a clear, accountable evidence workbench without losing the brand's energy or the reader's focus.
        OWN-WORLD: Evidence Workbench — quiet paper/graphite surfaces, status-first registers, magenta-to-purple action cues, and mono wayfinding labels.
        STORY: Counselor reads and reviews; Stakeholder fills and supports with evidence.
        FIRST VIEWPORT: Project status, posture metrics, next action, owner/attention context, and the Function register are visible before detailed outcomes.
        FORM: Evidence Workbench, candidate 3 of the grounded direction set, seed 250b41bb; Versotis Trust Workspace remains the user-pinned brand reference. FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md.
        */}
        {children}
      </body>
    </html>
  );
}
