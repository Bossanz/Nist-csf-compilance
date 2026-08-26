"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { APIError, api } from "../../lib/api";
import { LoginForm } from "../../components/LoginForm";

export default function LoginPage() {
  const router = useRouter();
  const [authChecked, setAuthChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void api.me()
      .then(() => router.replace("/organizations"))
      .catch((cause) => {
        if (active && (!(cause instanceof APIError) || cause.status !== 401)) {
          setError(messageOf(cause));
        }
      })
      .finally(() => {
        if (active) setAuthChecked(true);
      });
    return () => { active = false; };
  }, [router]);

  async function login(input: { email: string; password: string }) {
    setLoading(true);
    setError("");
    try {
      await api.login(input);
      router.replace("/organizations");
    } catch (cause) {
      setError(messageOf(cause));
    } finally {
      setLoading(false);
    }
  }

  if (!authChecked) return <main className="screen-center" role="status" aria-live="polite" aria-busy="true">Loading…</main>;
  return <LoginForm loading={loading} error={error} onSubmit={login} />;
}

function messageOf(cause: unknown) {
  return cause instanceof Error ? cause.message : "Something went wrong";
}
