"use client";

type Props = {
  loading: boolean;
  error: string;
  submitted: boolean;
  onSubmit: (email: string) => void;
};

export function ForgotPasswordForm({ loading, error, submitted, onSubmit }: Props) {
  return (
    <section className="auth-panel" aria-labelledby="forgot-password-title">
      <h1 id="forgot-password-title">Forgot password?</h1>
      <p className="muted">Enter your account email and we’ll send a reset link if the account is active.</p>
      {error && <div className="error" role="alert">{error}</div>}
      {submitted ? (
        <div className="panel auth-message" role="status">If an active account exists, a password reset link will be sent.</div>
      ) : (
        <form className="auth-form" onSubmit={(event) => { event.preventDefault(); const email = new FormData(event.currentTarget).get("email"); onSubmit(String(email || "").trim().toLowerCase()); }}>
          <label className="field"><span>Email</span><input name="email" type="email" autoComplete="email" required /></label>
          <button className="primary" type="submit" disabled={loading}>{loading ? "Sending…" : "Send reset link"}</button>
        </form>
      )}
      <a className="auth-link" href="/login">Back to sign in</a>
    </section>
  );
}
