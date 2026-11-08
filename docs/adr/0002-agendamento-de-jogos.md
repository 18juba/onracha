# ADR 0002: Horários pontuais e agendamento de jogos

- Status: Aceito
- Data: 2026-11-08

## Contexto

Uma arena precisa disponibilizar horários específicos para jogos. Um horário não pode receber mais de um jogo, e um usuário não pode participar de jogos com períodos sobrepostos.

## Decisão

- O proprietário cadastra intervalos pontuais com início e fim; horários recorrentes ficam fora desta primeira versão.
- Cada jogo referencia um horário, e o banco impede mais de um jogo no mesmo horário.
- Quem agenda o jogo entra automaticamente como participante.
- O agendamento e a confirmação de presença são transacionais. A operação bloqueia a linha do usuário enquanto verifica conflitos de intervalo e grava a participação, evitando reservas simultâneas conflitantes.

## Consequências

O modelo mantém a agenda simples e permite várias participações do mesmo usuário em horários diferentes. A validação de conflitos compara intervalos e usa transações do PostgreSQL para evitar dupla reserva concorrente.
