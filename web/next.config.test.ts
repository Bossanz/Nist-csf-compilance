import { expect, test } from "vitest";
import nextConfig from "./next.config";

test("proxies API routes to the internal Go service", async () => {
  const rewrites = await nextConfig.rewrites!();

  expect(rewrites).toContainEqual({
    source: "/api/:path*",
    destination: "http://localhost:8080/api/:path*",
  });
});
