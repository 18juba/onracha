# Guia de estilo visual

Este guia reúne as decisões visuais já combinadas para o ONRACHA. Use-o como referência ao criar telas e componentes; atualize-o quando uma decisão de produto mudar.

## Direção visual

- Esportiva, abstrata e moderna, com uma interface simples e visual.
- Priorize o conteúdo principal e ações diretas. Evite eyebrows, textos de apoio desnecessários, aparência de painel administrativo e decoração com estética de jogo.
- Use ícones e ilustrações para dar personalidade sem competir com o conteúdo.

## Layout e navegação

- Projete primeiro para dispositivos móveis.
- Em telas maiores, mantenha a aplicação em um container central estreito, com largura máxima atual de `430px`. Não expanda a interface para preencher o desktop.
- Respeite as áreas seguras do dispositivo, especialmente na parte inferior.
- A navegação principal fica na parte inferior e tem quatro destinos, nesta ordem: **Buscar**, **Jogos**, **Arenas** e **Perfil**.
- A bottombar ocupa sua própria área de layout. O conteúdo, incluindo o futuro mapa, deve caber acima dela sem ficar encoberto.
- Mantenha botões e alvos de toque confortáveis para uso com uma mão; use pelo menos `44px` como referência mínima.

## Cores e temas

O modo escuro usa um fundo roxo quase preto e superfícies escuras. O modo claro usa fundo branco e superfícies claras. O ciano `#09b0f4` é a cor de destaque nos dois temas.

Os tokens globais em `src/routes/layout.css` são a fonte de verdade para cores. Prefira esses tokens a valores de cor escritos diretamente nos componentes.

| Token                | Escuro                  | Claro                    |
| -------------------- | ----------------------- | ------------------------ |
| `--page-outside`     | `#09070f`               | `#edf0f3`                |
| `--app-background`   | `#15121e`               | `#fff`                   |
| `--surface`          | `rgb(36 33 45 / 82%)`   | `rgb(246 249 250 / 92%)` |
| `--surface-raised`   | `#25212e`               | `#f2f6f7`                |
| `--field-background` | `rgb(27 24 36 / 86%)`   | `#f5f8f9`                |
| `--text-primary`     | `#f8f7fb`               | `#171923`                |
| `--text-secondary`   | `#aaa6b6`               | `#5e6571`                |
| `--text-muted`       | `#817c8e`               | `#7a828e`                |
| `--accent`           | `#09b0f4`               | `#09b0f4`                |
| `--accent-hover`     | `#30baf6`               | `#30baf6`                |
| `--accent-ink`       | `#06272c`               | `#06272c`                |
| `--border`           | `rgb(255 255 255 / 9%)` | `rgb(20 31 43 / 10%)`    |
| `--focus-ring`       | `rgb(9 176 244 / 24%)`  | `rgb(9 176 244 / 24%)`   |

Sucesso e erro usam os tokens semânticos `--success` e `--danger`. Componentes e mapas devem responder ao tema ativo; não mantenha uma paleta separada que ignore a escolha da pessoa.

## Tipografia e conteúdo

- Use Inter com a pilha de fontes do sistema como fallback, conforme definido globalmente.
- Crie hierarquia com poucos níveis claros de tamanho e peso.
- Prefira rótulos curtos que indiquem a ação ou o destino. Inclua texto de apoio apenas quando ele ajudar a concluir a tarefa.

## Ícones, ilustrações e movimento

- Use os ícones do pacote `@lucide/svelte` para manter um traço consistente.
- Ícones decorativos devem ser ocultos da tecnologia assistiva. Botões que dependem apenas de um ícone precisam de um nome acessível.
- Use animações breves para transições e resposta a ações; elas não devem atrasar o fluxo nem carregar informação essencial.
- Respeite `prefers-reduced-motion` e ofereça estados compreensíveis sem animação.

## Mapa

- O mapa deve ser limpo e minimalista: ruas principais, formas de bairros e referências suficientes para a pessoa se localizar.
- Evite excesso de detalhes geográficos. As arenas devem se destacar por marcadores claros na cor de destaque.
- O mapa acompanha o tema claro ou escuro do aplicativo.
- A alternância entre mapa e lista deve preservar o espaço da bottombar e os controles essenciais para uso com uma mão.

## Implementação

- Mantenha `:root`, `body` e os tokens de tema em `src/routes/layout.css`.
- Use classes Tailwind para estilos locais dos componentes. Extraia um componente quando um padrão visual ou de interação for reutilizado.
- Evite regras globais para resolver estilos que pertencem a um componente específico.

## Revisão visual rápida

Antes de concluir uma tela, confira se ela funciona em viewport móvel e no container central de desktop, nos temas claro e escuro, com a bottombar e as áreas seguras. Verifique também contraste, foco visível, rótulos acessíveis e movimento reduzido.
