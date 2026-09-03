import type { Metadata } from "next";
import { GuestOnly } from "@/components/auth/guest-only";
import { LoginForm } from "@/components/auth/login-form";

export const metadata: Metadata = {
  title: "Entrar | JudCalc",
  description: "Entre na sua conta JudCalc.",
};

export default function LoginPage() {
  return (
    <GuestOnly>
      <LoginForm />
    </GuestOnly>
  );
}
