"use client";

import { useState } from "react";

type Props = { loading:boolean; error:string; onSubmit:(input:{email:string;password:string})=>void };

export function LoginForm({loading,error,onSubmit}:Props){
  const [email,setEmail]=useState("");const [password,setPassword]=useState("");
  return <main className="auth-layout"><section className="auth-intro"><span className="section-index">NIST CSF 2.0 / ACCESS</span><h1>Compliance, organized.</h1><p>One workspace for counselors and customer stakeholders to assess, review, and improve cybersecurity outcomes.</p><div className="auth-rule"><strong>106</strong><span>CSF outcomes ready for assessment</span></div></section><section className="auth-panel" aria-labelledby="sign-in-title"><div><span className="section-index">SECURE SESSION</span><h2 id="sign-in-title">Sign in</h2><p className="muted">Use the account created by your counselor or organization admin.</p></div>{error&&<div className="error" role="alert">{error}</div>}<form className="auth-form" onSubmit={event=>{event.preventDefault();onSubmit({email:email.trim().toLowerCase(),password})}}><label className="field"><span>Email</span><input type="email" autoComplete="email" required value={email} onChange={event=>setEmail(event.target.value)}/></label><label className="field"><span>Password</span><input type="password" autoComplete="current-password" required value={password} onChange={event=>setPassword(event.target.value)}/></label><button className="primary" disabled={loading}>{loading?"Signing in…":"Sign in"}</button></form></section></main>;
}
