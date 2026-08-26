import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, expect, test } from "vitest";
import { ThemeToggle } from "./ThemeToggle";

beforeEach(() => {
  window.localStorage.clear();
  document.documentElement.removeAttribute("data-theme");
});

test("switches between QID dark and light themes", () => {
  render(<ThemeToggle />);

  const toggle = screen.getByRole("button", { name: /enable light theme/i });
  expect(toggle.getAttribute("aria-pressed")).toBe("true");

  fireEvent.click(toggle);

  expect(document.documentElement.getAttribute("data-theme")).toBe("light");
  expect(window.localStorage.getItem("csf-theme")).toBe("light");
  expect(screen.getByRole("button", { name: /enable dark theme/i }).getAttribute("aria-pressed")).toBe("false");
});
