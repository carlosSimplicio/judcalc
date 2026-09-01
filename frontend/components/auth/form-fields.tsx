"use client";

import { useState, type InputHTMLAttributes } from "react";

type TextFieldProps = InputHTMLAttributes<HTMLInputElement> & {
  id: string;
  label: string;
};

export function TextField({ id, label, ...props }: TextFieldProps) {
  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      <input id={id} {...props} />
    </div>
  );
}

type PasswordFieldProps = TextFieldProps & {
  helpText?: string;
};

export function PasswordField({
  id,
  label,
  helpText,
  ...props
}: PasswordFieldProps) {
  const [visible, setVisible] = useState(false);
  const helpId = helpText ? `${id}-help` : undefined;

  return (
    <div className="field">
      <label htmlFor={id}>{label}</label>
      <div className="password-wrap">
        <input
          id={id}
          type={visible ? "text" : "password"}
          aria-describedby={helpId}
          {...props}
        />
        <button
          type="button"
          aria-controls={id}
          aria-pressed={visible}
          onClick={() => setVisible((current) => !current)}
        >
          {visible ? "Ocultar" : "Mostrar"}
        </button>
      </div>
      {helpText ? <p id={helpId} className="field-help">{helpText}</p> : null}
    </div>
  );
}

export function PasswordStrength({ password }: { password: string }) {
  const score = [
    password.length >= 8,
    /\p{L}/u.test(password),
    /\p{N}/u.test(password),
    /[^\p{L}\p{N}]/u.test(password),
  ].filter(Boolean).length;
  const level = !password
    ? ""
    : score <= 2
      ? "weak"
      : score === 3
        ? "medium"
        : "strong";
  const labels: Record<string, string> = {
    "": "não informada",
    weak: "fraca",
    medium: "média",
    strong: "forte",
  };

  return (
    <div className="password-strength">
      <div
        className={`password-meter${level ? ` is-${level}` : ""}`}
        aria-hidden="true"
      >
        <span />
      </div>
      <p aria-live="polite">Segurança da senha: {labels[level]}</p>
    </div>
  );
}
