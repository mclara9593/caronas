package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	protocolo "github.com/mclara9593/caronas/internal/protocol"
)

// função para conectar ao servidor com tentativas limitadas
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

// esqueleto do cliente
type Cliente struct {
	Nome   string
	Conn   net.Conn
	Reader *bufio.Reader // Leitor do socket para ler as respostas do servidor
}

// Construtor do Cliente
func newCliente(nome string, conn net.Conn) *Cliente {
	return &Cliente{
		Nome:   nome,
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

// EnviarRequisicao envia qualquer ação/payload e retorna a resposta do servidor
func (c *Cliente) EnviarRequisicao(action string, payload interface{}) (*protocolo.Response, error) {
	// 1. Serializa o payload específico para bytes de JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %v", err)
	}

	// Monta o envelope genérico da requisição
	req := protocolo.Request{
		Action:    action,
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), // ID único por timestamp
		Payload:   payloadBytes,
	}

	// 3. Serializa o envelope completo para JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// 4. Envia pelo socket TCP com o delimitador '\n' para enquadramento [Kurose]
	_, err = c.Conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar dados pelo socket: %v", err)
	}

	// 5. Aguarda e lê a resposta vinda do servidor até o '\n'
	respostaBytes, err := c.Reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta do servidor: %v", err)
	}

	// 6. Decodifica a resposta no struct Response
	var resp protocolo.Response
	err = json.Unmarshal(respostaBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	return &resp, nil
}

// 1. Troque 'conn net.Conn' por 'cliente *Cliente'
func autenticarPassageiro(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	fmt.Printf("Autenticando passageiro: %s...\n", user)
}

// 2. Troque 'conn net.Conn' por 'cliente *Cliente'
func buscarItinerarios(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Cidade de Origem: ")
	origem, _ := reader.ReadString('\n')
	origem = strings.TrimSpace(origem)

	fmt.Print("Cidade de Destino: ")
	destino, _ := reader.ReadString('\n')
	destino = strings.TrimSpace(destino)

	fmt.Print("Data desejada (ex: 2026-09-10): ")
	data, _ := reader.ReadString('\n')
	data = strings.TrimSpace(data)

	fmt.Printf("Buscando opções de viagens de %s para %s no dia %s...\n", origem, destino, data)
}

// 3. Troque 'conn net.Conn' por 'cliente *Cliente'
func reservarItinerario(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID do itinerário/trechos que deseja reservar: ")
	itinerarioID, _ := reader.ReadString('\n')
	itinerarioID = strings.TrimSpace(itinerarioID)

	fmt.Printf("Iniciando transação atômica de reserva para o itinerário %s...\n", itinerarioID)
}

// 4. Troque 'conn net.Conn' por 'cliente *Cliente'
func consultarReservas(cliente *Cliente) {
	fmt.Println("Buscando reservas ativas do usuário...")
}

// 5. Troque 'conn net.Conn' por 'cliente *Cliente'
func cancelarReserva(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID da Reserva que deseja cancelar: ")
	reservaID, _ := reader.ReadString('\n')
	reservaID = strings.TrimSpace(reservaID)

	fmt.Printf("Solicitando cancelamento da reserva %s...\n", reservaID)
}

func main() {
	// 1. Tenta conectar ao Servidor Central na porta 12000
	conn, err := conectarAoServidor("localhost:12000", 5)
	if err != nil {
		fmt.Printf("❌ Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// 2. Cria a instância do cliente
	cliente := newCliente("Passageiro", conn)
	fmt.Println("✔ Conectado com sucesso ao Servidor Central do VaiJunto!")

	// 3. Executa o loop do menu
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- VAIJUNTO: CLIENTE PASSAGEIRO ---")
		fmt.Println("1. Autenticar-se (Login)")
		fmt.Println("2. Buscar Itinerários de Carona")
		fmt.Println("3. Reservar Itinerário (Confirmação Atômica)")
		fmt.Println("4. Consultar Minhas Reservas")
		fmt.Println("5. Cancelar Reserva")
		fmt.Println("6. Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			autenticarPassageiro(cliente, reader)
		case "2":
			buscarItinerarios(cliente, reader)
		case "3":
			reservarItinerario(cliente, reader)
		case "4":
			consultarReservas(cliente)
		case "5":
			cancelarReserva(cliente, reader)
		case "6":
			fmt.Println("Encerrando conexão do passageiro...")
			return
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}
