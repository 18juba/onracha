# Frontend

Construído com SvelteKit, Svelte 5 e TypeScript. O projeto usa Bun como gerenciador de pacotes.

## Requisitos

- Bun 1.4.4, a mesma versão usada no Docker Compose.

## Instalar dependências

A partir da raiz do repositório:

```bash
cd frontend
bun install --frozen-lockfile
```

## Executar em desenvolvimento

```bash
bun run dev
```

Abra <http://localhost:5173>.

## Verificações e testes

```bash
bun run check
bun run lint
bun run test:unit -- --run
```

`check` executa a verificação de tipos e diagnósticos do Svelte; `lint` verifica formatação e ESLint; `test:unit -- --run` executa os testes unitários do Vitest uma vez; `bun run test:unit` inicia o Vitest em modo interativo. No scaffold inicial, o teste de exemplo está em `src/lib/vitest-examples/greet.spec.ts`.

## Build de produção e preview local

```bash
bun run build
bun run preview
```