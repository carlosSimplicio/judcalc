# JudCalc Frontend

Frontend do JudCalc criado com Next.js, React e TypeScript.

## Requisitos

- Node.js 24 LTS
- npm

## Desenvolvimento

```bash
nvm use
npm install
cp .env.example .env.local
npm run dev
```

A aplicação estará disponível em `http://localhost:3000`.

As páginas de autenticação estão disponíveis em `/login` e `/cadastro`. Por
padrão, elas acessam a API em `http://localhost:8080/api/v1`. Altere
`NEXT_PUBLIC_API_BASE_URL` em `.env.local` quando o backend estiver em outro
endereço:

```dotenv
NEXT_PUBLIC_API_BASE_URL=http://localhost:8081/api/v1
```

O backend deve permitir a origem do frontend em `CORS_ALLOWED_ORIGINS`.
