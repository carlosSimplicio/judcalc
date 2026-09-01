"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState, type FormEvent } from "react";
import { authErrorToast } from "@/lib/auth/error-toast";
import { signUp } from "@/lib/auth/client";
import { saveSession } from "@/lib/auth/session";
import { Button } from "@/components/ui/button";
import { useToast } from "@/components/ui/toast-provider";
import { AuthShell } from "./auth-shell";
import { PasswordField, PasswordStrength, TextField } from "./form-fields";

export function SignupForm() {
  const router = useRouter();
  const { showToast } = useToast();
  const [password, setPassword] = useState("");
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
      const session = await signUp({
        name: String(data.get("name") ?? ""),
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
      titleId="signup-title"
      title="Crie sua conta"
      description="Comece a precificar com mais segurança."
      footer={
        <>
          Já tem uma conta? <Link href="/login">Entre aqui</Link>
        </>
      }
    >
      <form onSubmit={handleSubmit}>
        <TextField
          id="signup-name"
          name="name"
          label="Nome completo"
          type="text"
          autoComplete="name"
          required
          disabled={submitting}
        />
        <TextField
          id="signup-email"
          name="email"
          label="E-mail"
          type="email"
          autoComplete="email"
          required
          disabled={submitting}
        />
        <PasswordField
          id="signup-password"
          name="password"
          label="Senha"
          autoComplete="new-password"
          minLength={8}
          maxLength={72}
          required
          disabled={submitting}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
          helpText="Use ao menos 8 caracteres."
        />
        <PasswordStrength password={password} />
        <div className="field checkbox-field">
          <label>
            <input name="terms" type="checkbox" required disabled={submitting} />
            <span>Concordo com os Termos de Uso e a Política de Privacidade</span>
          </label>
        </div>
        <Button
          type="submit"
          loading={submitting}
          loadingLabel="Criando conta..."
        >
          Criar conta
        </Button>
      </form>
    </AuthShell>
  );
}
