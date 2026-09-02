# JudCalc

API para consulta da tabela de honorários da OAB-SP. O projeto separa o backend
Go da preparação de dados em Python:

```text
backend/  API Go, testes e banco SQLite gerado
python/   scripts, testes, esquema, PDF e JSON fonte
docs/     documentação do produto
```

## Requisitos

- Go 1.24 ou superior
- Python 3 com as dependências de `python/requirements.txt`

## Preparar o banco

O arquivo `python/data/oab-sp.json` é a fonte de verdade. A partir da raiz do
projeto, instale as dependências e crie ou sincronize o banco regenerável:

```powershell
python -m pip install -r python/requirements.txt
python -m python.scripts.init_database
```

O comando grava `backend/data/app.db`.

O projeto ainda não possui migrações. Após uma alteração incompatível no
esquema, apague esse arquivo e execute o comando novamente.

## Executar a API

```powershell
Set-Location backend
go mod download
go run ./cmd/api
```

A API escuta em `:8080` e abre `data/app.db` para leitura e escrita. As
configurações podem ser substituídas por variáveis de ambiente:

```powershell
$env:HTTP_ADDR = ":8081"
$env:DATABASE_PATH = "C:\caminho\app.db"
$env:CORS_ALLOWED_ORIGINS = "http://localhost:3000,http://127.0.0.1:3000"
Set-Location backend
go run ./cmd/api
```

`CORS_ALLOWED_ORIGINS` aceita uma lista de origens separadas por vírgulas. Se
não for informada, somente `http://localhost:3000` será permitida. As origens
devem ser completas, incluindo protocolo e porta.

Se o banco não existir ou tiver esquema incompatível, a aplicação encerra na
inicialização com uma mensagem de erro.

## Endpoints

- `POST /api/v1/auth/sign-up`
- `POST /api/v1/auth/sign-in`
- `GET /api/v1/areas`
- `GET /api/v1/services`
- `GET /api/v1/fixed-costs`
- `PATCH /api/v1/fixed-costs`
- `POST /api/v1/services/fee-calculation`
- `GET /healthz` (health check público)

Cadastre um usuário informando email, nome e uma senha de 8 a 72 bytes:

```json
{
  "email": "advogada@example.com",
  "name": "Maria Silva",
  "password": "uma-senha-segura"
}
```

O cadastro e o sign-in retornam um token opaco com validade de 30 dias:

```json
{
  "data": {
    "user": {
      "id": 1,
      "email": "advogada@example.com",
      "name": "Maria Silva"
    },
    "access_token": "token-retornado-pela-api",
    "token_type": "Bearer",
    "expires_at": "2026-10-01T12:00:00Z"
  }
}
```

Para entrar novamente, envie `email` e `password` para
`POST /api/v1/auth/sign-in`. Todas as outras rotas exigem o token:

```text
Authorization: Bearer token-retornado-pela-api
```

Corpos JSON são limitados a 64 KiB.

Os endpoints de áreas e serviços aceitam `page` (padrão `1`), `page_size` (padrão `20`, máximo `100`) e
`q`. A busca ignora acentos, busca prefixos e exige correspondência de todos os
termos.

Exemplo:

```text
GET /api/v1/services?page=1&page_size=20&q=acao%20previd
```

Os custos fixos pertencem ao usuário identificado pelo token. Quando ele ainda
não possui custos cadastrados, o GET retorna todas as categorias zeradas:

```text
GET /api/v1/fixed-costs
```

O PATCH cria o registro ou atualiza somente os custos enviados. Os valores são
inteiros em centavos; a anuidade da OAB é convertida para uma média mensal na
resposta.

```json
{
  "costs": {
    "oab_annual_fee": { "annual_amount_cents": 120000 },
    "internet": { "monthly_amount_cents": 15000 }
  }
}
```

### Cálculo de honorários

O cálculo usa os custos fixos do usuário autenticado e mantém a referência da
OAB separada dos valores econômicos. Envie horas positivas e os níveis `low`,
`medium` ou `high`:

```json
{
  "service_id": 42,
  "estimated_hours": 10,
  "billable_hours_per_month": 80,
  "complexity": "medium",
  "risk": "high"
}
```

```text
POST /api/v1/services/fee-calculation
```

Complexidade usa fatores `1.00`, `1.25` e `1.50`; risco usa `1.00`, `1.10` e
`1.20`. A resposta apresenta a referência da OAB, os fatores usados, o custo
operacional por hora, o custo mínimo sustentável e a estimativa técnica.

Quando não há custos fixos utilizáveis, a referência da OAB continua presente,
os resultados econômicos são `null` e `warnings` contém
`fixed_costs_unavailable`.

## Testes

```powershell
Set-Location backend
go test ./...
go vet ./...
Set-Location ..
python -m unittest discover -s python/tests
```

## Deploy na VPS

O workflow [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml) roda os
testes, empacota o commit e envia o código para a VPS por SCP. Em seguida, acessa
a VPS por SSH, constrói as imagens Docker no próprio servidor e atualiza os
containers. Nenhuma imagem da aplicação é publicada em um registry. O fluxo é
executado em pushes para `main` e também pode ser iniciado manualmente em
**Actions**.

### Preparar a VPS

Instale Docker Engine com o plugin Docker Compose e permita que o usuário de
deploy execute `docker` sem `sudo`. Depois crie `~/judcalc/.env` para esse
usuário:

```dotenv
CORS_ALLOWED_ORIGINS=https://app.exemplo.com
NEXT_PUBLIC_API_BASE_URL=https://app.exemplo.com/api/v1
HOST_BIND_ADDRESS=0.0.0.0
HTTP_PORT=80
```

O Nginx do Compose é o único container que publica uma porta no host. Ele
encaminha `/api/*` para o backend e as demais rotas para o frontend, conforme
[`infra/nginx.conf`](infra/nginx.conf). Para colocar outro proxy responsável por
HTTPS à frente dele, use `HOST_BIND_ADDRESS=127.0.0.1` e uma porta livre em
`HTTP_PORT`. Ao expor diretamente em `0.0.0.0`, configure o firewall da VPS.

O banco fica no volume nomeado `judcalc_backend_data`. A cada inicialização, o
backend sincroniza áreas e serviços a partir do JSON incluído na imagem sem
remover usuários, tokens ou custos fixos.

Cada commit enviado é extraído em `~/judcalc/releases/<sha>`. Releases antigas
podem ser removidas manualmente depois que o novo deploy for verificado; as
imagens e o volume do banco não dependem desses diretórios após o build.

### Configurar o GitHub

Crie um environment chamado `production` e cadastre nele estes secrets:

- `VPS_HOST`: hostname ou IP da VPS;
- `VPS_USER`: usuário SSH preparado para o deploy;
- `VPS_SSH_PRIVATE_KEY`: chave privada dedicada ao workflow;
- `VPS_SSH_KNOWN_HOSTS`: linha verificada do host, obtida por exemplo com
  `ssh-keyscan -H seu-host` e conferida contra a fingerprint da VPS;
- `VPS_SSH_PORT`: porta SSH, opcional; o padrão é `22`.

O usuário SSH precisa ter acesso de escrita a `~/judcalc` e permissão para
executar Docker. A URL pública da API é lida do `.env` da VPS e incorporada ao
frontend durante o build.
