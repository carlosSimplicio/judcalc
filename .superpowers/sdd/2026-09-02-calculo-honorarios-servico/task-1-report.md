# Task 1 — cálculo puro de honorários

## Implementação

- Adicionados `FeeLevel`, fatores de complexidade/risco, entrada/saída do cálculo e `ErrInvalidFeeCalculation`.
- Adicionada `ValidateFeeCalculationInput`, reutilizada por `CalculateFee`, validando horas positivas e finitas e níveis conhecidos.
- Adicionadas consolidação mensal de custos (`MonthlyAverageCents` e `FixedCosts.MonthlyTotalCents`) e cálculo com precisão integral antes do arredondamento em centavos.
- Handler de custos fixos passou a consumir `domain.MonthlyAverageCents`, removendo a implementação privada duplicada.

## Arquivos alterados

- `backend/internal/domain/fixed_costs.go`
- `backend/internal/domain/fee_calculation.go`
- `backend/internal/domain/fee_calculation_test.go`
- `backend/internal/httpapi/handlers/fixed_costs.go`

## TDD: evidência RED/GREEN

Os testes foram escritos antes da implementação. A primeira execução focada com o shim `go` retornou código 159 sem executar o compilador (o shim asdf local está sem runtime funcional); isso foi uma falha de ambiente, não uma aprovação falsa. Após disponibilizar o binário Go instalado diretamente e um `GOCACHE` gravável em `/tmp`, os testes focados passaram após a implementação:

```text
ok github.com/carlosSimplicio/judcalc/backend/internal/domain 0.002s
```

O conjunto de testes cobre o contrato ausente (tipos/funções inexistentes na fase RED), consolidação, fatores, precisão, custos zero, níveis inválidos e a validação pública de valores não finitos.

## Comandos e resultados

- `gofmt -w ...` — OK.
- `git diff --check` — OK.
- `go test ./internal/domain -run 'TestMonthlyTotal|TestCalculateFee|TestValidateFeeCalculationInput'` — PASS.
- `go test ./internal/domain ./internal/httpapi/handlers ./...` — PASS.
- `go test ./internal/httpapi -run 'TestGetFixedCosts|TestPatchFixedCosts'` — PASS.

Os comandos foram executados com o Go 1.25.1 instalado localmente e `GOCACHE`/`GOPATH` em `/tmp`; foi necessária autorização de rede uma vez para baixar dependências ausentes.

## Self-review

- Fórmulas usam `float64` somente no cálculo intermediário e `math.Round` uma vez por resultado.
- A anuidade é arredondada por resto (`>= 6`) e todas as categorias de custo são somadas.
- Custos mensais zero retornam ponteiros econômicos nulos.
- Não há alteração de esquema, persistência ou comportamento de patch de custos.
- `git diff --check` não apontou problemas.

## Preocupações

- O shim asdf `go` do ambiente retorna 159; a verificação foi feita pelo binário Go 1.25.1 diretamente.
- A execução inicial sem rede não conseguiu baixar dependências; a suíte completa passou após a autorização.
