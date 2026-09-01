import { ApiError } from "./client";

type ErrorToast = {
  title: string;
  message: string;
};

const ERROR_TITLES: Record<string, string> = {
  invalid_credentials: "Credenciais inválidas",
  email_already_exists: "E-mail já cadastrado",
  invalid_body: "Dados inválidos",
  request_too_large: "Solicitação inválida",
  internal_error: "Erro no servidor",
  network_error: "Falha de conexão",
  unexpected_response: "Erro inesperado",
};

export function authErrorToast(error: unknown): ErrorToast {
  if (error instanceof ApiError) {
    return {
      title: ERROR_TITLES[error.code] ?? "Não foi possível continuar",
      message: error.message,
    };
  }

  return {
    title: "Erro inesperado",
    message: "Não foi possível concluir a solicitação. Tente novamente.",
  };
}
