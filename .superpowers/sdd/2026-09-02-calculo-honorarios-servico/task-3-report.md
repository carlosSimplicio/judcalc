# Task 3 — endpoint autenticado de cálculo por serviço

## Estado herdado

- Foram encontrados quatro arquivos não commitados: `router.go`, `router_test.go`, `fee_calculation_test.go` e `handlers/fee_calculation.go`.
- Na primeira execução focada disponível, os testes já estavam verdes. Não há registro verificável de uma execução RED do agente interrompido, portanto não é possível afirmar que a implementação herdada seguiu TDD.
- Não havia comportamento ausente depois da inspeção: o handler já validava o corpo antes de acessar repositórios e os testes cobriam essa regra. Assim, não foi adicionada produção nova nem fabricada uma evidência RED retroativa.

## Implementação preservada e revisada

- Adicionado `POST /api/v1/services/fee-calculation` ao grupo autenticado.
- O handler decodifica JSON estrito, rejeita `service_id <= 0`, chama `domain.ValidateFeeCalculationInput` e retorna `400 invalid_body` antes de consultar qualquer repositório para entradas semanticamente inválidas.
- Consulta o serviço pelo `service_id` do corpo, busca custos pelo usuário autenticado, calcula os valores e expõe serviço, referência OAB, entradas/fatores, cálculo e avisos.
- Mapeia serviço ausente para `404 service_not_found`; falhas internas não expõem detalhes; custos zero retornam valores econômicos nulos e o aviso `fixed_costs_unavailable`.

## Arquivos

- `backend/internal/httpapi/handlers/fee_calculation.go`
- `backend/internal/httpapi/fee_calculation_test.go`
- `backend/internal/httpapi/router.go`
- `backend/internal/httpapi/router_test.go`

## Testes e verificações

- O shim `go` do asdf retornou código 159 e não foi usado como evidência de teste.
- `/home/csimplicio/.asdf/installs/golang/1.25.1/go/bin/go test -count=1 -v ./internal/httpapi -run '^TestFeeCalculation'` — PASS (quatro testes e todos os subtestes).
- `/home/csimplicio/.asdf/installs/golang/1.25.1/go/bin/go test -count=1 ./...` — PASS: `auth`, `domain`, `httpapi` e `storage/sqlite`; os demais pacotes não possuem testes.
- `gofmt -w` nos quatro arquivos e `git diff --check` — OK.
- As execuções usaram `TMPDIR`, `GOCACHE` e `GOMODCACHE` graváveis em `/dev/shm`; dependências declaradas foram baixadas após autorização.

## Evidência TDD e qualidade dos testes

- RED histórico: indisponível. Os testes e a produção já existiam quando esta retomada começou, e a primeira execução focada observável foi PASS.
- GREEN verificável: o teste focado acima passou contra a rota real e seus handlers, cobrindo cálculo completo, resposta sem custos, corpo inválido sem chamadas aos repositórios e mapeamentos de erro.
- As expectativas de valores usam fixtures literais e os stubs apenas delimitam repositórios; asserções observam contrato HTTP e identificadores recebidos pelos repositórios.

## Self-review e preocupações

- Revisados a rota exata, autenticação, uso exclusivo de `service_id` no corpo, ordem de validação, nulos JSON, ausência de vazamento de erro e aviso de custos indisponíveis.
- Não há mudanças em esquema, carga ou FTS; `AGENTS.md` não requer atualização.
- A única preocupação operacional é o shim Go quebrado; use o binário Go 1.25.1 explícito com caches temporários para verificações futuras.
