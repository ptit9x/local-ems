import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { restoreSession } from "./api";
import { DashboardShell } from "./components/dashboard/dashboard-shell";
import { LoginForm } from "./features/auth/components/login-form";

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
    <LoginForm done={() => setSignedIn(true)} />
  );
}
