# git-repo-rewind

> Uma máquina do tempo visual e explorável para qualquer repositório git — direto no terminal.

`rewind` transforma a história de um repositório git numa experiência **animada e navegável**.
Em vez de um replay passivo, você controla um **cursor de tempo** e percorre livremente a
evolução do projeto, vendo o mesmo instante através de quatro lentes: timeline de commits,
heatmap de atividade, árvore de branches e evolução das linguagens.

Feito em Go com o ecossistema [Charm](https://charm.sh) (Bubble Tea, Lipgloss, Harmonica),
[`go-git`](https://github.com/go-git/go-git) para a extração e
[`go-enry`](https://github.com/go-enry/go-enry) para detecção de linguagens. Funciona 100%
offline, em qualquer repo local, sem configuração.

---

## Instalação

Requer **Go 1.26+**.

```sh
go install github.com/JoaoVictorVM/git-repo-rewind/cmd/rewind@latest
```

Ou compilando a partir do código:

```sh
git clone https://github.com/JoaoVictorVM/git-repo-rewind
cd git-repo-rewind
go build -o rewind ./cmd/rewind
```

## Uso

Dentro de um repositório git:

```sh
rewind
```

Ou apontando para outro diretório:

```sh
rewind -path /caminho/para/o/repo
rewind -version
```

O extrator percorre a história uma vez ao abrir; depois disso a navegação é instantânea.

### Atalhos

| Tecla | Ação |
|---|---|
| `h` / `l` (ou `←` / `→`) | Move o cursor para trás / para frente |
| `g` / `G` | Pula para o início / fim da história |
| `+` / `-` | Alterna a granularidade do passo (commit → dia → semana) |
| `space` | Play/pause do autoplay a partir do cursor |
| `Tab` (ou `1`–`4`) | Troca de cena |
| `o` | Modo overview (grid 2×2 com as quatro cenas) |
| `s` | Card de resumo ("printável") |
| `Shift+T` | Troca de tema |
| `?` | Mostra/esconde a ajuda |
| `q` / `Ctrl+C` | Sair |

## Cenas

Todas compartilham o mesmo cursor de tempo — mover o cursor atualiza a cena para o estado
acumulado da história **até aquele instante**. Os valores não pulam: deslizam com easing.

- **Timeline** — gráfico de área de linhas adicionadas (rosa) e removidas (ciano) ao longo do
  tempo, contadores animados e feed de log dos commits recentes.
- **Heatmap** — grade dia da semana × hora que acende conforme o cursor avança, revelando
  padrões (madrugadas, fins de semana, *crunch time*).
- **Branches** — fluxo de commits com merges destacados e estatísticas do DAG (commits,
  merges, ramos abertos).
- **Linguagens** — barras horizontais com a composição de linguagens no instante do cursor,
  reajustando conforme a história muda.

Além delas: **overview** (`o`) mostra as quatro simultaneamente num grid, e o **card de resumo**
(`s`) traz o frame de fecho — dia mais produtivo, maior sequência, arquivo mais tocado, mix de
linguagens e totais.

## Temas

As cores são **tokens de tema**, nunca fixas no código. Troque em runtime com `Shift+T`:

- `default` — paleta sóbria (rosa / azul / roxo sobre fundo escuro).
- `nerv` — vermelho, âmbar e verde, inspirado nas telas de alerta de Evangelion.

## Arquitetura

O princípio central é **event sourcing visual**, com três camadas fortemente desacopladas.
Isso permite trocar a fonte de dados sem tocar na animação e garante replay determinístico
(mesma entrada → mesmo resultado).

```
┌─────────────┐   stream de     ┌──────────────┐   estado do    ┌────────────┐
│  EXTRATOR   │─── eventos ───▶ │    MOTOR     │── mundo por ──▶│  RENDERER  │
│ (go-git)    │  normalizado    │ (world state)│    cursor       │ (Bubble Tea)│
└─────────────┘                 └──────────────┘                 └────────────┘
```

- **`internal/extract`** — lê o `.git` via `go-git` atrás da interface `EventSource` e emite um
  stream de eventos normalizado, ordenado por timestamp. A detecção de linguagem (`go-enry`)
  acontece aqui.
- **`internal/engine`** — consome o stream e mantém o `WorldState` para qualquer posição do
  cursor. É **puro** (sem I/O, sem rendering): indexa por timestamp e usa busca binária +
  *prefix sums* para scrub instantâneo. 100% testável.
- **`internal/tui`** — app Bubble Tea; só desenha o `WorldState` e traduz teclas em comandos.
  Lipgloss para estilo, Harmonica para o easing dos valores animados.
- **`internal/theme`** — mapa de tokens → cores. As cenas pedem cor por token, nunca por hex.
- **`internal/stats`** — cálculo puro do card de resumo.

Regra de ouro: o renderer nunca fala com o git direto, e o motor nunca renderiza. Adicionar
uma cena nova = consumir tipos de evento existentes (ou introduzir um novo), sem tocar nas
outras camadas.

## Desenvolvimento

```sh
go test ./...        # motor e stats têm prioridade de cobertura
go build ./...
go vet ./...
```

## Créditos

Inspirado no [`github-visualize`](https://github.com/akitaonrails/github-visualize) de
Fabio Akita, que faz replay animado de repositórios num dashboard web. `git-repo-rewind`
leva a ideia para o terminal e troca o replay passivo por navegação interativa.

## Licença

MIT.
