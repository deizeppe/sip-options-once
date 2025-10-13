# sip-options-once

Envia **um** pacote SIP **OPTIONS** para um ou mais destinos e imprime o resultado da resposta.  
Focado em ser **simples**, **rápido** e **portável** (macOS/Apple Silicon e Linux), com controle explícito do **IP/porta de origem**.

---

## Sumário

- [Recursos](#recursos)
- [Requisitos](#requisitos)
- [Instalação do Go (macOS Apple Silicon)](#instalação-do-go-macos-apple-silicon)
- [Criando o projeto](#criando-o-projeto)
- [Compilação](#compilação)
- [Uso (CLI)](#uso-cli)
- [Como funciona](#como-funciona)
- [Dicas e troubleshooting](#dicas-e-troubleshooting)
- [Notas de segurança](#notas-de-segurança)
- [Licença](#licença)

---

## Recursos

- Envia **1** SIP `OPTIONS` por IP e **aguarda 1 resposta** (sem repetir, sem quarentena).
- Faz bind no **IP/porta de origem**; tenta *fallback* para `0.0.0.0:porta` se o IP não existir na máquina, e por fim deixa o SO escolher a porta.
- Valida:
  - Status-line `SIP/2.0 200 ...`
  - `Call-ID` de resposta igual ao enviado
- Sem dependências externas além da stdlib do Go (usa `github.com/google/uuid` para gerar Call-ID).

---

## Requisitos

- **Go 1.21+** (recomendado 1.22+)  
- A máquina onde roda o binário precisa possuir o **IP de origem** na interface de rede (se quiser garantir o “source IP” exato).

---

## Instalação do Go (macOS Apple Silicon)

**Homebrew (recomendado):**
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew install go
go version
```

**Manual (.pkg):**
1. Baixe em https://go.dev/dl/ (macOS arm64).
2. Instale o `.pkg`.
3. Adicione ao PATH:
   ```bash
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
   exec zsh
   ```

---

## Criando o projeto

```bash
cd sip-options-once
go mod init sip-options-once
go get github.com/google/uuid
```

Cole o código do `main.go` e salve.

---

## Compilação

```bash
go build -trimpath -ldflags="-s -w" -o sip-options-once
```

Cross-compile:
```bash
GOOS=linux  GOARCH=amd64 go build -o sip-options-once-linux-amd64
GOOS=linux  GOARCH=arm64 go build -o sip-options-once-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o sip-options-once-macos-amd64
```

Com Makefile:
```bash
make build
make all
make clean
```

---

## Uso (CLI)

### Exemplo

```bash
./sip-options-once   -ips "177.70.70.55"   -src-ip 145.135.777.75   -src-port 5060   -kam-port 5060   -timeout 2
```

### Parâmetros

| Flag        | Obrigatório | Descrição | Padrão |
|--------------|-------------|-----------|---------|
| `-ips` | Sim | Lista de IPs destino separada por vírgula | — |
| `-src-ip` | Sim | IP de origem para o pacote UDP | — |
| `-src-port` | Sim | Porta de origem UDP | — |
| `-kam-port` | Não | Porta SIP do destino | 5060 |
| `-timeout` | Não | Timeout em segundos | 2 |
| `-sip-user` | Não | Usuário SIP (From/Contact) | SIPMonitor |
| `-to-user` | Não | Usuário SIP (To) | SIPMonitor |

### Saídas

```
177.70.70.55 - 200 OK (Call-ID OK)
177.70.70.55 - TIMEOUT
177.70.70.55 - unexpected: SIP/2.0 404 Not Found
```

---

## Como funciona

1. Gera `Call-ID` e `branch` únicos.
2. Constrói o pacote SIP OPTIONS e envia via UDP.
3. Aguarda uma resposta e imprime o resultado.

---

## Dicas e troubleshooting

- **Verificar IP de origem:**  
  `ifconfig | grep 145.135.777.75` (macOS)  
  `ip addr | grep 145.135.777.75` (Linux)

- **Capturar tráfego:**  
  ```bash
  sudo tcpdump -n -i any udp and port 5060 or host 177.70.70.55
  ```

- **Portas privilegiadas:** Use portas >1024.

- **Firewall:** valide ACLs e rotas.

---

## Notas de segurança

- Nenhuma autenticação SIP é usada.
- Não requer privilégios administrativos.
- Recomendado uso apenas em redes controladas.

---

### Exemplo rápido

```bash
./sip-options-once   -ips "177.70.70.55"   -src-ip 145.135.777.75   -src-port 5060   -kam-port 5060   -timeout 2
```


## 📝 Licença

MIT License.

---

## 👨‍💻 Autor

Ricardo Deizeppe — Engenharia VoIP  
Baseado no dialer original em Perl, reimplementado em **Python + FastAPI**.
