## Requisito

- Go 1.27.3, conforme declarado em `go.mod`.

## Executar localmente

A partir da raiz do repositório:

```bash
cd backend
go run .
```

A API fica disponível em <http://localhost:8080>. A rota `GET /` retorna uma mensagem JSON.

## Testes

Para executar os testes do backend:

```bash
go test ./...
```