import type { ButtonHTMLAttributes } from "react";

type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  loading?: boolean;
  loadingLabel?: string;
};

export function Button({
  children,
  className = "",
  disabled,
  loading = false,
  loadingLabel = "Aguarde...",
  ...props
}: ButtonProps) {
  return (
    <button
      className={`button ${className}`.trim()}
      disabled={disabled || loading}
      aria-busy={loading}
      {...props}
    >
      {loading ? loadingLabel : children}
    </button>
  );
}
