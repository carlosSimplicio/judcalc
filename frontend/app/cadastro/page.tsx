import type { Metadata } from "next";
import { GuestOnly } from "@/components/auth/guest-only";
import { SignupForm } from "@/components/auth/signup-form";

export const metadata: Metadata = {
  title: "Criar conta | JudCalc",
  description: "Crie sua conta JudCalc.",
};

export default function SignupPage() {
  return (
    <GuestOnly>
      <SignupForm />
    </GuestOnly>
  );
}
