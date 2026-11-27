# plugin starter — Go

Plugin de exemplo mínimo para a [plugin platform](https://github.com/giomartinsdev/discord-plugin-platform). Conta quantas vezes
cada membro usou `!contar`. Só a biblioteca padrão.

## Rodar

```bash
git clone https://github.com/giomartinsdev/discord-plugin-platform
git clone https://github.com/giomartinsdev/plugin-starter-go
cd discord-plugin-platform

node core/cli.js conformance ../plugin-starter-go
node core/cli.js install ../plugin-starter-go
node core/cli.js simulate contador contar

# build estático, se preferir não usar `go run`:
cd ../plugin-starter-go && go build -o contador .
# e troque cmd no plugin.toml por ["./contador"]
```

## Comandos de teste

| comando | resposta esperada |
|---|---|
| `!contar` | `<@1> chegou a **1**.` e, a cada 10, mais `10 é um número redondo!` |
| `!contagem` | `Você usou !contar **1** vez(es).` |
| `!contagem` (primeira vez) | `Você nunca usou !contar. Comece agora.` |
| `!outro` | `Comando outro não existe aqui.` |

## O contrato

- `stdout`: JSON-RPC 2.0, uma mensagem por linha. Nada mais — em Go é fácil esquecer um `fmt.Println`.
- `stderr`: log livre.
- Chamadas ao core são síncronas: o mesmo `bufio.Scanner` lê as respostas.
- O plugin nunca recebe o token do Discord.

O spec completo do protocolo (`PROTOCOL.md`) fica no repositório da plataforma, `giomartinsdev/discord-plugin-platform` (privado).

## `go run` no manifesto

`cmd = ["go", "run", "."]` compila a cada início (~1 s). Para produção, `go build` no `runtime.build`
e `cmd = ["./contador"]`. O `timeout_ms` do manifesto do exemplo é 8000 por causa da compilação.
