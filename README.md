# ONRACHA

Plataforma em desenvolvimento para cadastrar arenas, organizar e encontrar jogos para participar.

## Fluxo principal

```mermaid
flowchart LR
    A[Acessar ONRACHA] --> B[Entrar ou criar conta]
    B --> C{O que deseja fazer?}

    C -->|Quero organizar| D[Cadastrar arena]
    D --> E[Organizar um racha]

    C -->|Quero jogar| F[Encontrar um racha]

    E --> G[Participar da partida]
    F --> G
```

## Executar com Docker Compose

Copie `.env.example` para `.env` se quiser personalizar os valores locais. Os valores padrão são apenas para desenvolvimento e não devem ser usados em produção.

Na raiz do repositório, execute:

```bash
docker compose up
```

O Compose inicia PostgreSQL, aplica migrations pendentes, inicia a API e inicia o frontend:

- Frontend: <http://localhost:5173>
- API: <http://localhost:8080>
- Swagger UI: <http://localhost:8080/swagger>
- OpenAPI: <http://localhost:8080/openapi.yaml>
- PostgreSQL: `localhost:5432`

Para encerrar, pressione `Ctrl+C`. Se os serviços foram iniciados em segundo plano, use:

```bash
docker compose down
```

Os volumes `postgres_data` e `arena_photos` mantêm os dados e as fotos locais. `docker compose down -v` também apaga esses volumes.

## Smoke test

O smoke test com k6 percorre cadastro e login de duas contas, criação de arena e horário, agendamento, busca do jogo, participação e consulta a sala de jogo. Cada execução cria registros permanentes no banco configurado; rode manualmente apenas contra uma base de teste descartável.

Com a API e as migrations disponíveis, execute:

    k6 run performance/smoke.js

O teste roda automaticamente no job de backend da CI, usando um PostgreSQL temporário do job. O k6 está fixado em 2.3.0 na workflow.

## Documentação

- [Backend e testes](backend/README.md)
- [Guia de contribuição e TDD](CONTRIBUTING.md)
- [Especificação OpenAPI](docs/openapi.yaml)
- [ADR](docs/adr)
