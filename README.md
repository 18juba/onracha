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

Na raiz do repositório, execute:

```bash
docker compose up
```

O Compose inicia os três serviços:

- Frontend: <http://localhost:5173>
- Backend: <http://localhost:8080>
- PostgreSQL: `localhost:5432`

Para encerrar, pressione `Ctrl+C`. Se os serviços foram iniciados em segundo plano, use:

```bash
docker compose down
```