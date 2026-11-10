# Guia de estilo visual

Este guia define a identidade visual do ON Racha e orienta as telas e componentes compartilhados.

## Direção visual

- Visual moderno, simples e inspirado na linguagem de componentes do shadcn: superfícies neutras, bordas discretas e ações ghost como padrão secundário.
- Reserve o acento verde-azulado dessaturado para a ação principal, seleção e estados ativos. Evite grandes áreas de cor saturada.
- Mostre primeiro a tarefa principal. Use rótulos curtos, hierarquia tipográfica clara e pouca decoração.
- A aplicação é mobile-first. Em telas maiores, mantenha o container central com largura máxima de `430px`.
- Preserve áreas seguras do dispositivo e os alvos de toque de pelo menos `44px`.

## Cores e temas

Os tokens globais em `src/routes/layout.css` são a fonte de verdade para os temas escuro e claro. Prefira-os a valores de cor escritos diretamente nos componentes.

| Token                | Escuro                  | Claro                    |
| -------------------- | ----------------------- | ------------------------ |
| `--page-outside`     | `#0b0c0f`               | `#f3f4f5`                |
| `--app-background`   | `#111317`               | `#ffffff`                |
| `--surface`          | `rgb(24 27 32 / 88%)`   | `rgb(255 255 255 / 90%)` |
| `--surface-raised`   | `#1b1f24`               | `#f5f6f7`                |
| `--field-background` | `#14171b`               | `#fafafa`                |
| `--text-primary`     | `#f3f4f4`               | `#191c20`                |
| `--text-secondary`   | `#a4aaad`               | `#626a71`                |
| `--text-muted`       | `#727a7d`               | `#858d94`                |
| `--accent`           | `#0785b5`               | `#09b0f4`                |
| `--border`           | `rgb(255 255 255 / 8%)` | `rgb(25 30 36 / 10%)`    |

Erros e sucesso usam `--danger`, `--danger-surface`, `--danger-surface-hover` e `--success`. Todos os controles devem continuar legíveis nos dois temas.
O texto e os ícones sobre superfícies primárias usam sempre branco (`--accent-ink`); elementos ativos sem fundo, como a navegação, usam `--accent`.

## Componentes e interação

- Reutilize os componentes em `src/lib/components/ui` para botões, campos, selects, datas, horários, checkbox e switch. Não crie estilos de controle isolados em telas.
- Use `primary` apenas para a ação principal da etapa; prefira `ghost` para ações auxiliares e `outline` para controles de seleção.
- Mantenha rótulos associados aos campos, foco visível, nomes acessíveis para ações de ícone e feedback de erro junto ao campo relevante.
- Selects, datepickers e timepickers usam painéis compactos com abertura breve. A seleção de checkbox/switch deve responder com uma transição discreta.
- Respeite `prefers-reduced-motion`; nenhuma informação pode depender da animação.
- Use os ícones de `@lucide/svelte` com o mesmo traço e o mesmo alinhamento visual.

## Implementação

- Use Svelte 5, TypeScript e classes Tailwind. Extraia um componente quando um padrão visual ou de interação for reutilizado.
- Mantenha temas e tokens em `src/routes/layout.css`; evite folhas globais para estilos que pertencem a um componente específico.
- Preserve a lógica e a semântica nativas de formulários, validação e estados desabilitados ao substituir controles.
- Atualize este guia quando uma decisão visual de produto mudar.
