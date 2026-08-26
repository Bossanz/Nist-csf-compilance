"use client";

import { useEffect, useState } from "react";

type Theme = "light" | "dark";

const STORAGE_KEY = "csf-theme";

type Props = {
  compact?: boolean;
};

function storedTheme(): Theme | null {
  try {
    const value = window.localStorage.getItem(STORAGE_KEY);
    return value === "light" || value === "dark" ? value : null;
  } catch {
    return null;
  }
}

function applyTheme(theme: Theme) {
  document.documentElement.dataset.theme = theme;
  document.documentElement.style.colorScheme = theme;
}

export function ThemeToggle({ compact = false }: Props) {
  const [theme, setTheme] = useState<Theme>("dark");

  useEffect(() => {
    const initialTheme = storedTheme() ?? "dark";
    setTheme(initialTheme);
    applyTheme(initialTheme);
  }, []);

  function toggleTheme() {
    const nextTheme: Theme = theme === "dark" ? "light" : "dark";
    setTheme(nextTheme);
    applyTheme(nextTheme);
    try {
      window.localStorage.setItem(STORAGE_KEY, nextTheme);
    } catch {
      // The theme still applies when storage is unavailable.
    }
  }

  const isDark = theme === "dark";
  return (
    <button
      className="theme-toggle"
      type="button"
      aria-label={isDark ? "Enable light theme" : "Enable dark theme"}
      aria-pressed={isDark}
      title={isDark ? "Switch to light theme" : "Switch to dark theme"}
      onClick={toggleTheme}
    >
      {compact ? (
        <span className="theme-toggle-icon" aria-hidden="true">
          <svg viewBox="0 0 20 20">
            {isDark ? <path d="M14.8 13.8A6.8 6.8 0 0 1 6.2 5.2 6.8 6.8 0 1 0 14.8 13.8Z" /> : <><circle cx="10" cy="10" r="3.2" /><path d="M10 2v2M10 16v2M2 10h2M16 10h2M4.3 4.3l1.4 1.4M14.3 14.3l1.4 1.4M15.7 4.3l-1.4 1.4M5.7 14.3l-1.4 1.4" /></>}
          </svg>
        </span>
      ) : (
        <>
          <span className="theme-toggle-label">{isDark ? "Light" : "Dark"}</span>
          <span className="theme-toggle-status">Theme</span>
        </>
      )}
    </button>
  );
}
