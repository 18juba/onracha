## Fluxo de desenvolvimento

Para cada mudança:

1. Descreva o comportamento esperado e a regra de negócio envolvida.
2. Escreva um teste focado que represente esse comportamento.
3. Execute o teste e confirme que ele falha pelo motivo esperado (Red).
4. Implemente o mínimo necessário para fazê-lo passar (Green).
5. Refatore mantendo os testes verdes (Refactor).
6. Execute as verificações relevantes do projeto antes de considerar a mudança pronta.

Em correções de defeitos, primeiro crie um teste que reproduza o problema. Mantenha cada etapa pequena e evite incluir funcionalidades não relacionadas.

## Backend Go

- Mantenha regras de negócio no domínio, sem dependência de HTTP, banco de dados ou logging.
- Deixe handlers HTTP cuidarem da entrada e saída da API e chamarem a lógica da aplicação.
- Aplique SOLID de forma pragmática: funções e tipos com responsabilidades claras e dependências explícitas. Evite interfaces, camadas e padrões sem uma necessidade concreta.
- Se uma decisão arquitetural trouxer uma escolha relevante ou difícil de reverter, registre-a em um ADR.

### Testes

- Testes unitários do domínio devem cobrir regras e resultados relevantes sem iniciar servidor ou banco de dados.
- Testes de integração devem verificar que componentes reais funcionam juntos em uma fronteira importante, como roteamento e resposta HTTP. Para HTTP, prefira `net/http/httptest`.
- Use nomes de testes que expliquem o cenário e o resultado esperado.
- O mínimo é **80% de cobertura nos pacotes de domínio**. A CI aplica esse limite a `internal/auth`; ao adicionar outros pacotes de domínio, inclua-os na verificação. Cobertura complementa os testes de comportamento; não substitui bons cenários.

### Logs estruturados

- Use `log/slog` da biblioteca padrão e configure o logger da aplicação com saída JSON.
- Registre eventos nas bordas da aplicação, como inicialização e requisições HTTP, incluindo campos úteis como método, rota, status, duração e erro.
- Mantenha o domínio independente do logger; retorne erros para a camada responsável por registrá-los.
- Nunca registre senhas, tokens, segredos, corpos de requisição ou dados pessoais sem necessidade.

### Verificações locais

Na pasta `backend/`, execute:

```bash
gofmt -w caminho/do/arquivo.go
go vet ./...
go test ./...
go build ./...
```

A CI também verifica a formatação com `gofmt -l .`.

## Frontend

Na pasta `frontend/`, use Bun e os scripts definidos em `package.json`:

```bash
bun install --frozen-lockfile
bun run check
bun run lint
bun run test:unit --run
bun run build
```

Para mudanças em Svelte ou SvelteKit, siga também as instruções de `frontend/AGENTS.md`.

## Executar a aplicação

Na raiz do repositório, use Docker Compose. O serviço `migrate` aplica as migrations depois que o PostgreSQL fica saudável e antes de iniciar a API:

```bash
docker compose up
```

Para executar apenas os testes de domínio sem banco:

```bash
docker compose run --rm --no-deps -e DATABASE_URL= backend go test ./internal/auth/...
```

Para executar também o teste de integração com PostgreSQL:

```bash
docker compose run --rm backend go test ./...
```

Atualize `docs/openapi.yaml` quando alterar os endpoints públicos e registre decisões arquiteturais relevantes em `docs/adr/`.

## Mensagens de commit

Use o formato Conventional Commits:

```text
<tipo>(<escopo opcional>): <descrição curta>
```

Tipos comuns: `feat`, `fix`, `test`, `refactor`, `docs`, `ci` e `chore`.
