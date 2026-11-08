# ADR-0001: Arquitetura inicial do backend e persistência

- **Status:** Aceito
- **Data:** 2026-11-07

## Contexto

A aplicação deverá lidar com contas, arenas, jogos e participantes. A estrutura deve permitir testar regras de negócio sem exigir o banco e continuar simples para o tamanho atual do projeto.

## Decisão

- Manter a API como um monólito modular: um único serviço, organizado em pacotes internos por funcionalidade, como `auth`, `arena` e `game`, quando cada funcionalidade for implementada.
- Separar as responsabilidades de HTTP, regras/casos de uso e persistência. O ponto de inicialização da API conecta essas partes.
- Usar o padrão Repository nas funcionalidades que precisam persistir dados. A interface será pequena e ficará junto do pacote que a consome; a implementação PostgreSQL ficará na camada de persistência.
- Usar PostgreSQL 18.6 em Docker Compose para desenvolvimento local, com volume nomeado e health check. A porta do banco será publicada apenas em `127.0.0.1`.
- Modelar inicialmente a relação como um usuário responsável por várias arenas: cada arena tem `owner_user_id` obrigatório com chave estrangeira para `users`; a exclusão do usuário é restrita enquanto houver arenas vinculadas. Co-gestores ficam fora do escopo inicial.
- Versionar mudanças de estrutura de dados com `golang-migrate`. O serviço `migrate` do Compose aplica as migrations pendentes antes de iniciar a API.
- Armazenar senhas com Argon2id e emitir sessões opacas em cookie `HttpOnly`. O banco persiste apenas o hash SHA-256 do token de sessão, com expiração de 24 horas.
- Manter credenciais reais em `.env`, ignorado pelo Git. `.env.example` e os valores padrão do Compose são exemplos locais fictícios e não devem ser usados em produção.
- Manter o banco de produção fora do Compose de desenvolvimento, preferencialmente como serviço gerenciado.

## Justificativa do Repository

As regras de negócio não devem executar SQL nem depender do driver do banco. O Repository deixa explícito o que cada caso de uso precisa persistir e permite testá-lo com uma implementação em memória, sem iniciar PostgreSQL.

Essa escolha é adequada porque os dados do ONRACHA serão persistentes e relacionados. Ela não garante que trocar de banco será automático: consultas, migrações e diferenças entre bancos continuam exigindo trabalho.

## Relação com SOLID

- **Responsabilidade única:** handlers HTTP cuidam do protocolo; casos de uso aplicam as regras; o adaptador PostgreSQL cuida do armazenamento.
- **Inversão de dependência:** os casos de uso dependem da interface Repository, não do PostgreSQL.
- **Segregação de interfaces:** cada Repository expõe apenas as operações que seu consumidor precisa.

Os demais princípios serão aplicados quando forem úteis, sem criar camadas ou abstrações apenas para cumprir uma lista.

## Consequências

- Os testes unitários das regras e casos de uso de autenticação rodam sem banco de dados.
- O teste integrado percorre cadastro, login e consulta da conta usando PostgreSQL real quando `DATABASE_URL` estiver definida; a CI inicia PostgreSQL e aplica as migrations antes dos testes.
- O Compose inicia PostgreSQL, aplica as migrations e então inicia a API e o frontend.
- O usuário pode ter várias arenas; cada arena possui um único usuário responsável nesta primeira versão.
- O volume `postgres_data` preserva os dados entre inicializações. `docker compose down -v` remove esse volume e os dados locais.
- Senhas e tokens de sessão nunca são armazenados em texto puro. Tokens de sessão não são enviados no JSON da API.
- A imagem PostgreSQL está fixada em 18.6; atualizações de versão serão feitas de forma deliberada.

## Referências

- [Política de versionamento do PostgreSQL](https://www.postgresql.org/support/versioning/)
- [Imagem oficial do PostgreSQL no Docker Hub](https://hub.docker.com/_/postgres)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
