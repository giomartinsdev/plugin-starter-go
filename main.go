// Plugin de exemplo em Go — contador por membro.
//
// Protocolo: JSON-RPC 2.0, uma mensagem JSON por linha, em stdin/stdout.
// Só a biblioteca padrão: encoding/json, bufio, os.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type rpcMsg struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type params struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	UserID  string   `json:"userId"`
	GuildID string   `json:"guildId"`
}

var (
	in  = bufio.NewScanner(os.Stdin)
	seq = 0
)

func send(v any) {
	b, err := json.Marshal(v)
	if err != nil {
		logf("falha ao serializar: %v", err)
		return
	}
	os.Stdout.Write(append(b, '\n'))
}

func logf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[info] "+format+"\n", args...)
}

func reply(id json.RawMessage, result any) {
	send(map[string]any{"jsonrpc": "2.0", "id": id, "result": result})
}

func fail(id json.RawMessage, code int, message string) {
	send(map[string]any{"jsonrpc": "2.0", "id": id, "error": map[string]any{"code": code, "message": message}})
}

// core chama o core de forma síncrona: escreve o pedido e lê até achar
// a resposta com o nosso id. Chamadas do core nesse meio-tempo recebem ok.
func core(method string, p map[string]any) (map[string]any, error) {
	seq++
	id := fmt.Sprintf("p%d", seq)
	send(map[string]any{"jsonrpc": "2.0", "id": id, "method": method, "params": p})

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg rpcMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			logf("linha inválida: %.60s", line)
			continue
		}
		if msg.Method != "" {
			reply(msg.ID, map[string]any{"ok": true})
			continue
		}
		var gotID string
		json.Unmarshal(msg.ID, &gotID)
		if gotID != id {
			continue
		}
		if msg.Error != nil {
			return nil, fmt.Errorf("%s", msg.Error.Message)
		}
		var out map[string]any
		json.Unmarshal(msg.Result, &out)
		return out, nil
	}
	return nil, fmt.Errorf("core fechou a conexão")
}

func handleCommand(p params) []map[string]any {
	key := "count:" + p.UserID

	switch p.Command {
	case "contar":
		res, err := core("kv.get", map[string]any{"guildId": p.GuildID, "key": key})
		if err != nil {
			return []map[string]any{{"type": "reply", "content": "não consegui ler o contador agora"}}
		}
		total := 1
		if v, ok := res["value"].(float64); ok {
			total = int(v) + 1
		}
		if _, err := core("kv.set", map[string]any{"guildId": p.GuildID, "key": key, "value": total}); err != nil {
			return []map[string]any{{"type": "reply", "content": "não consegui salvar o contador agora"}}
		}
		milestone := ""
		if total%10 == 0 {
			milestone = fmt.Sprintf("\n%d é um número redondo!", total)
		}
		return []map[string]any{{
			"type":    "reply",
			"content": fmt.Sprintf("<@%s> chegou a **%d**.%s", p.UserID, total, milestone),
		}}

	case "contagem":
		res, err := core("kv.get", map[string]any{"guildId": p.GuildID, "key": key})
		if err != nil {
			return []map[string]any{{"type": "reply", "content": "não consegui ler o contador agora"}}
		}
		total := 0
		if v, ok := res["value"].(float64); ok {
			total = int(v)
		}
		if total == 0 {
			return []map[string]any{{"type": "reply", "content": "Você nunca usou `!contar`. Comece agora."}}
		}
		return []map[string]any{{
			"type":    "reply",
			"content": fmt.Sprintf("Você usou `!contar` **%d** vez(es).", total),
		}}
	}

	return []map[string]any{{"type": "reply", "content": fmt.Sprintf("Comando `%s` não existe aqui.", p.Command)}}
}

func onRequest(msg rpcMsg) {
	var p params
	json.Unmarshal(msg.Params, &p)

	switch msg.Method {
	case "initialize":
		var init struct {
			Guilds []struct {
				ID string `json:"id"`
			} `json:"guilds"`
		}
		json.Unmarshal(msg.Params, &init)
		logf("inicializado para %d servidor(es)", len(init.Guilds))
		reply(msg.ID, map[string]any{"ok": true})

	case "handleCommand":
		reply(msg.ID, handleCommand(p))

	case "handleEvent":
		reply(msg.ID, []map[string]any{})

	case "health":
		reply(msg.ID, map[string]any{"ok": true, "detail": "contador ativo"})

	case "shutdown":
		reply(msg.ID, map[string]any{"ok": true})
		os.Exit(0)

	default:
		fail(msg.ID, -32601, "método desconhecido: "+msg.Method)
	}
}

func main() {
	in.Buffer(make([]byte, 0, 1<<20), 1<<20)
	logf("contador iniciado")

	for in.Scan() {
		line := strings.TrimSpace(in.Text())
		if line == "" {
			continue
		}
		var msg rpcMsg
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			logf("linha inválida: %.60s", line)
			continue
		}
		if msg.Method == "" {
			continue // resposta para uma chamada que já foi resolvida
		}
		onRequest(msg)
	}
}
