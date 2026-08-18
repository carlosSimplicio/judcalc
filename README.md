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
Set-Location backend
go run ./cmd/api
```

Se o banco não existir ou tiver esquema incompatível, a aplicação encerra na
inicialização com uma mensagem de erro.

## Endpoints

- `GET /api/v1/areas`
- `GET /api/v1/services`
- `GET /api/v1/fixed-costs/:user_id`
- `PATCH /api/v1/fixed-costs`

Os endpoints de áreas e serviços aceitam `page` (padrão `1`), `page_size` (padrão `20`, máximo `100`) e
`q`. A busca ignora acentos, busca prefixos e exige correspondência de todos os
termos.

Exemplo:

```text
GET /api/v1/services?page=1&page_size=20&q=acao%20previd
```

Os custos fixos são consultados pelo identificador textual do usuário. Quando
ainda não há cadastro, o GET retorna todas as categorias zeradas:

```text
GET /api/v1/fixed-costs/user-123
```

O PATCH cria o registro ou atualiza somente os custos enviados. Os valores são
inteiros em centavos; a anuidade da OAB é convertida para uma média mensal na
resposta.

```json
{
  "user_id": "user-123",
  "costs": {
    "oab_annual_fee": { "annual_amount_cents": 120000 },
    "internet": { "monthly_amount_cents": 15000 }
  }
}
```

## Testes

```powershell
Set-Location backend
go test ./...
go vet ./...
Set-Location ..
python -m unittest discover -s python/tests
```
