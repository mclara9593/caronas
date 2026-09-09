package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

func conectarAoServidor(endereco string, maxTentativas int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 1; i <= maxTentativas; i++ {
		fmt.Printf("🔌 Tentando conectar ao Servidor Central (%d/%d)...\n", i, maxTentativas)

		// net.DialTimeout evita que o cliente fique travado indefinidamente se a rede falhar
		conn, err = net.DialTimeout("tcp", endereco, 3*time.Second)
		if err == nil {
			return conn, nil // Conexão estabelecida com sucesso!
		}

		fmt.Printf("⚠️ Falha na conexão: %v. Tentando novamente em 2 segundos...\n", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err // Retorna o último erro se esgotar as tentativas
}

func main() {
	// O cliente inicia a conexão TCP com o servidor informando o syn
	enderecoServidor := "localhost:12000"

	conn, err := conectarAoServidor(enderecoServidor, 5)
	if err != nil {
		fmt.Println("❌ Erro definitivo: Não foi possível contatar o Servidor Central.")
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("✔ Conectado com sucesso ao Servidor Central!")

	// Inicialização do leitor do socket...
	reader := bufio.NewReader(conn)
	inputReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Digite uma mensagem (ou 'sair'): ")
		text, _ := inputReader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "sair" {
			break
		}

		// Envia com quebra de linha para enquadramento correto no servidor
		conn.Write([]byte(text + "\n"))

		// Lê a resposta do servidor
		resposta, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao ler do servidor:", err)
			break
		}
		fmt.Printf("Resposta do servidor: %s", resposta)
	}
}
