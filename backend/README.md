## Executar com Docker Compose

Na raiz do repositório:

```bash
docker compose up
```

O Compose inicia o PostgreSQL, aplica as migrations pendentes e inicia a API. Endereços locais:

- API: <http://localhost:8080>
- Verificação de saúde: <http://localhost:8080/health>
- Documentação interativa Swagger UI: <http://localhost:8080/swagger>
- Especificação OpenAPI: <http://localhost:8080/openapi.yaml> (`docs/openapi.yaml` no repositório)

## Executar testes

Os testes de domínio não precisam de banco. Na raiz:

```bash
docker compose run --rm --no-deps -e DATABASE_URL= backend go test ./internal/auth/...
```

O teste de integração usa PostgreSQL.

```bash
docker compose run --rm backend go test ./...
```

## Verificações do backend

Na raiz do repositório:

```bash
docker compose run --rm --no-deps backend go vet ./...
docker compose run --rm --no-deps backend go build ./...
```