import { expect, test, vi } from "vitest";
import { redirect } from "next/navigation";
import Page from "./page";

vi.mock("next/navigation", () => ({ redirect: vi.fn() }));

test("redirects the root route to organizations", () => {
  Page();
  expect(redirect).toHaveBeenCalledWith("/organizations");
});
