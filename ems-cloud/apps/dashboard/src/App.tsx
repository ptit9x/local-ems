import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertTriangle,
  LoaderCircle,
  LockKeyhole,
  ShieldCheck,
} from "lucide-react";
import { login, restoreSession } from "./api";
import { DashboardShell } from "./components/dashboard/dashboard-shell";
import { Badge } from "./components/ui/badge";
import { Brand as Logo } from "./components/ui/brand";
import { Button } from "./components/ui/button";
import { Field, FieldLabel } from "./components/ui/field";
import { Input } from "./components/ui/input";

function Login({ done }: { done: () => void }) {
  const [email, setEmail] = useState("admin@demo.local");
  const [password, setPassword] = useState("demo1234");
  const queryClient = useQueryClient();
  const signIn = useMutation({
    mutationFn: () => login(email, password),
    onSuccess: () => {
      queryClient.setQueryData(["session"], true);
      done();
    },
  });
  return (
    <main className="min-h-screen bg-[#07110f] text-white lg:grid lg:grid-cols-[1.08fr_.92fr]">
      <section className="relative hidden overflow-hidden border-r border-white/10 p-10 lg:flex lg:flex-col xl:p-12">
        <div className="login-grid absolute inset-0 opacity-45" />
        <div className="absolute -left-24 top-24 size-96 rounded-full bg-emerald-400/15 blur-[100px]" />
        <div className="absolute bottom-0 right-0 size-[34rem] rounded-full bg-cyan-400/10 blur-[120px]" />
        <div className="relative z-10">
          <Logo />
        </div>
        <div className="relative z-10 my-auto max-w-2xl">
          <Badge className="mb-7 border-emerald-300/20 bg-emerald-300/10 px-3 py-1 text-emerald-200">
            <span className="mr-2 size-1.5 animate-pulse rounded-full bg-emerald-300" />
            Simulator environment
          </Badge>
          <h1 className="font-display text-5xl font-semibold leading-[1.08] tracking-[-.04em] xl:text-6xl">
            Energy intelligence,
            <br />
            <span className="text-emerald-300">without the noise.</span>
          </h1>
          <p className="mt-7 max-w-xl text-lg leading-8 text-slate-300">
            One operational view for storage, solar, grid and every connected
            asset. Built for teams who need clarity in real time.
          </p>
          <div className="mt-12 grid max-w-xl grid-cols-3 gap-4">
            {[
              ["37", "Connected devices"],
              ["99.98%", "Data availability"],
              ["5 sec", "Live refresh"],
            ].map(([value, label]) => (
              <div
                key={label}
                className="rounded-2xl border border-white/10 bg-white/[.045] p-4 backdrop-blur"
              >
                <div className="font-display text-2xl font-semibold">
                  {value}
                </div>
                <div className="mt-1 text-xs text-slate-400">{label}</div>
              </div>
            ))}
          </div>
        </div>
        <p className="relative z-10 text-xs text-slate-500">
          CDI EMS Platform · Secure simulator workspace
        </p>
      </section>
      <section className="flex min-h-screen items-center justify-center bg-[#f4f5f0] px-5 py-10 text-slate-950 sm:px-8">
        <div className="w-full max-w-[430px]">
          <div className="mb-10 flex lg:hidden">
            <Logo />
          </div>
          <div className="mb-8">
            <p className="mb-2 text-sm font-medium text-emerald-700">
              Welcome back
            </p>
            <h2 className="font-display text-3xl font-semibold tracking-[-.035em] sm:text-4xl">
              Sign in to your workspace
            </h2>
            <p className="mt-3 text-sm leading-6 text-slate-500">
              Monitor site health, energy flow and connected devices in one
              place.
            </p>
          </div>
          <form
            onSubmit={(event) => {
              event.preventDefault();
              signIn.mutate();
            }}
            className="space-y-5"
          >
            <Field>
              <FieldLabel htmlFor="login-email">Email address</FieldLabel>
              <Input
                id="login-email"
                autoComplete="email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                className="border-slate-200 bg-white text-slate-950"
              />
            </Field>
            <Field>
              <div className="flex items-center justify-between">
                <FieldLabel htmlFor="login-password">Password</FieldLabel>
                <button
                  type="button"
                  className="text-xs font-medium text-emerald-700 hover:underline"
                >
                  Forgot password?
                </button>
              </div>
              <Input
                id="login-password"
                type="password"
                autoComplete="current-password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                className="border-slate-200 bg-white text-slate-950"
              />
            </Field>
            {signIn.isError && (
              <div
                role="alert"
                className="flex gap-2 rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700"
              >
                <AlertTriangle className="mt-0.5 size-4 shrink-0" />
                Unable to sign in. Check your credentials and try again.
              </div>
            )}
            <Button
              type="submit"
              disabled={signIn.isPending}
              className="h-12 w-full bg-[#0c7664] text-white hover:bg-[#095f52]"
            >
              {signIn.isPending ? (
                <LoaderCircle className="size-4 animate-spin" />
              ) : (
                <LockKeyhole className="size-4" />
              )}
              Sign in securely
            </Button>
          </form>
          <div className="mt-7 flex items-center justify-center gap-2 text-center text-xs text-slate-500">
            <ShieldCheck className="size-4 shrink-0 text-emerald-700" />
            Protected by tenant-isolated access controls
          </div>
        </div>
      </section>
    </main>
  );
}

export function App() {
  const session = useQuery({
    queryKey: ["session"],
    queryFn: restoreSession,
    staleTime: Infinity,
    retry: false,
  });
  const [signedIn, setSignedIn] = useState(false);
  if (session.isLoading)
    return (
      <div className="grid min-h-screen place-items-center bg-[#07110f] text-emerald-200">
        <LoaderCircle className="size-6 animate-spin" />
      </div>
    );
  return session.data || signedIn ? (
    <DashboardShell signedOut={() => setSignedIn(false)} />
  ) : (
    <Login done={() => setSignedIn(true)} />
  );
}
