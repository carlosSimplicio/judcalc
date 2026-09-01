"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { authErrorToast } from "@/lib/auth/error-toast";
import { signIn } from "@/lib/auth/client";
import { saveSession } from "@/lib/auth/session";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { AuthShell } from "./auth-shell";
import { PasswordField, TextField } from "./form-fields";

export function LoginForm() {
  const router = useRouter();
  const { showToast } = useToast();
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (submitting) {
      return;
    }

    const form = event.currentTarget;
    if (!form.reportValidity()) {
      return;
    }

    const data = new FormData(form);
    setSubmitting(true);
    try {
      const session = await signIn({
        email: String(data.get("email") ?? ""),
        password: String(data.get("password") ?? ""),
      });
      saveSession(session);
      router.replace("/");
    } catch (error) {
      showToast(authErrorToast(error));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <AuthShell
      titleId="login-title"
      title="Entre na sua conta"
      description="Informe seus dados para continuar."
      footer={
        <>
          Novo por aqui? <Link href="/cadastro">Crie sua conta</Link>
        </>
      }
    >
      <form onSubmit={handleSubmit}>
        <TextField
          id="login-email"
          name="email"
          label="E-mail"
          type="email"
          autoComplete="email"
          required
          disabled={submitting}
        />
        <PasswordField
          id="login-password"
          name="password"
          label="Senha"
          autoComplete="current-password"
          required
          maxLength={72}
          disabled={submitting}
        />
        <Button
          type="submit"
          loading={submitting}
          loadingLabel="Entrando..."
        >
          Entrar
        </Button>
      </form>
    </AuthShell>
  );
}
