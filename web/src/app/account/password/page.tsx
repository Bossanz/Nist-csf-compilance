"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { APIError, api } from "../../../lib/api";
import { ChangePasswordForm } from "../../../components/ChangePasswordForm";
import { ThemeToggle } from "../../../components/ThemeToggle";

export default function ChangePasswordPage() {
  const router = useRouter();
  const [authChecked, setAuthChecked] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    void api.me()
      .then(() => { if (active) setAuthChecked(true); })
      .catch((cause) => {
        if (cause instanceof APIError && cause.status === 401) {
          router.replace("/login");
          return;
        }
        if (active) {
          setError(messageOf(cause));
          setAuthChecked(true);
        }
      });
    return () => { active = false; };
  }, [router]);

  async function changePassword(currentPassword: string, newPassword: string) {
    setLoading(true);
    setError("");
    try {
      await api.changePassword(currentPassword, newPassword);
      router.replace("/login");
    } catch (cause) {
      setError(messageOf(cause));
      setLoading(false);
    }
  }

  if (!authChecked) return <main className="screen-center">Loading account…</main>;
  return (
    <main className="auth-layout">
      <section className="auth-intro">
        <div className="auth-theme-control"><ThemeToggle /></div>
        <h1>Keep access secure.</h1>
        <p>Update your password whenever you need to. Active sessions will be revoked after the change.</p>
      </section>
      <ChangePasswordForm loading={loading} error={error} onSubmit={changePassword} />
    </main>
  );
}

function messageOf(cause: unknown) {
  return cause instanceof APIError ? cause.message : cause instanceof Error ? cause.message : "Something went wrong";
}
