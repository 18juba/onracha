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

- Os testes unitários de regras e casos de uso podem rodar sem banco de dados.
- O Compose inicia PostgreSQL junto dos serviços de frontend e backend; o backend ainda não se conecta ao banco nesta etapa.
- Ao implementar a primeira funcionalidade persistida, será necessário adicionar o driver, as migrações, a conexão e a composição do Repository.
- O volume `postgres_data` preserva os dados entre inicializações. `docker compose down -v` remove esse volume e os dados locais.
- A imagem está fixada em PostgreSQL 18.6; atualizações de versão serão feitas de forma deliberada.

## Referências

- [Política de versionamento do PostgreSQL](https://www.postgresql.org/support/versioning/)
- [Imagem oficial do PostgreSQL no Docker Hub](https://hub.docker.com/_/postgres)
