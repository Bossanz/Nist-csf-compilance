import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, expect, test } from "vitest";
import { ThemeToggle } from "./ThemeToggle";

beforeEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute("data-theme");
});

test("switches between Versotis light and dark themes", () => {
  render(<ThemeToggle />);

  const toggle = screen.getByRole("button", { name: /enable dark theme/i });
  expect(toggle.getAttribute("aria-pressed")).toBe("false");

  fireEvent.click(toggle);

  expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
  expect(window.localStorage.getItem("csf-theme")).toBe("dark");
  expect(screen.getByRole("button", { name: /enable light theme/i }).getAttribute("aria-pressed")).toBe("true");
});
