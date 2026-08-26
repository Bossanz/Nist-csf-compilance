import "./globals.css";
import type { Metadata } from "next";
export const metadata: Metadata = { title: "CSF Compliance", description: "NIST CSF 2.0 compliance workspace" };

const designContract = "QID v3 Compliance Workspace | dark-first workspace system | magenta-purple action gradient";

export default function RootLayout({ children }: Readonly<{ children: React.ReactNode }>) {
  return (
    <html lang="en" data-theme="dark" suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&family=Space+Grotesk:wght@400;500;600;700&display=swap"
          rel="stylesheet"
        />
      </head>
      <body data-design-contract={designContract}>
        {/*
        THESIS: Turn a complex compliance assessment into a clear, accountable workspace without losing the brand's energy.
        OWN-WORLD: QID v3 Compliance Workspace — dark graphite surfaces, magenta-to-purple action cues, and mono wayfinding labels.
        STORY: Counselor reads and reviews; Stakeholder fills and supports with evidence.
        FIRST VIEWPORT: Project context, Function index, current/target status, and one clear next action remain visible before the outcome list.
        FORM: Versotis Trust Workspace; user-pinned brand reference. FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md.
        */}
        {children}
      </body>
    </html>
  );
}
