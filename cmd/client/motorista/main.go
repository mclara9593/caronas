package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	protocol "github.com/mclara9593/caronas/internal/protocol"
)

// função para conectar ao servidor com tentativas limitadas
func conectarAoServidor(endereco string, maxTentativas int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 1; i <= maxTentativas; i++ {
		fmt.Printf(" Tentando conectar ao Servidor Central (%d/%d)...\n", i, maxTentativas)

		// net.DialTimeout evita que o cliente fique travado indefinidamente se a rede falhar
		conn, err = net.DialTimeout("tcp", endereco, 3*time.Second)
		if err == nil {
			return conn, nil // Conexão estabelecida com sucesso!
		}

		fmt.Printf("Falha na conexão: %v. Tentando novamente em 2 segundos...\n", err)
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
func (c *Cliente) EnviarRequisicao(action string, payload interface{}) (*protocol.Response, error) {
	// 1. Serializa o payload específico para bytes de JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %v", err)
	}

	// Monta o envelope genérico da requisição
	req := protocol.Request{
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
	var resp protocol.Response
	err = json.Unmarshal(respostaBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	return &resp, nil
}

func autenticarMotorista(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	fmt.Print("Digite sua senha: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	fmt.Print("Digite sua CNH: ")
	cnh, _ := reader.ReadString('\n')
	cnh = strings.TrimSpace(cnh)
	fmt.Printf("Autenticando Motorista: %s...\n", user)
}

func publicarCarona(cliente *Cliente, reader *bufio.Reader) {
	// 1. Coleta a rota completa (ex: Feira de Santana, Salvador, Aracaju)
	fmt.Print("Digite as cidades da rota (separadas por vírgula): ")
	rotaInput, _ := reader.ReadString('\n')
	rotaInput = strings.TrimSpace(rotaInput)

	// Converte a string em um slice []string
	cidadesRaw := strings.Split(rotaInput, ",")
	var rota []string
	for _, c := range cidadesRaw {
		cidadeTratada := strings.TrimSpace(c)
		if cidadeTratada != "" {
			rota = append(rota, cidadeTratada)
		}
	}

	if len(rota) < 2 {
		fmt.Println(" A rota precisa ter pelo menos 2 cidades (origem e destino).")
		return
	}

	// Coleta a data/horário de partida
	fmt.Print("Horário de partida (ex: 2026-09-12 18:00): ")
	depTime, _ := reader.ReadString('\n')
	depTime = strings.TrimSpace(depTime)

	// Coleta a capacidade de assentos
	fmt.Print("Quantidade de vagas disponíveis: ")
	var vagas int
	fmt.Scanln(&vagas)

	// Coleta o preço por trecho
	fmt.Print("Preço por trecho (R$): ")
	var preco float64
	fmt.Scanln(&preco)

	// Monta o payload conforme a struct PushRide
	payload := protocol.PushRide{
		Route:           rota,
		DepartureTime:   depTime,
		Capacity:        vagas,
		PricePerSegment: preco,
	}

	// Envia a requisição via socket
	resp, err := cliente.EnviarRequisicao("publish_ride", payload)
	if err != nil {
		fmt.Printf("Erro ao publicar carona: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func consultarCaronas(cliente *Cliente) {
	fmt.Println("Consultando caronas ativas do motorista...")
	// todo
}

func cancelarCarona(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID da carona que deseja cancelar: ")
	idCarona, _ := reader.ReadString('\n')
	idCarona = strings.TrimSpace(idCarona)

	// todo

	fmt.Printf("Cancelando carona com ID: %s...\n", idCarona)
}

func main() {
	//Tenta conectar ao Servidor Central
	conn, err := conectarAoServidor("localhost:9593", 5)
	if err != nil {
		fmt.Printf(" Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	//Cria a instância do cliente
	cliente := newCliente("Passageiro", conn)
	fmt.Println("Conectado com sucesso ao Servidor Central do VaiJunto!")

	// Executa o loop do menu
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- VAIJUNTO: CLIENTE MOTORISTA ---")
		fmt.Println("1. Autenticar-se (Login)")
		fmt.Println("2. Publicar Carona")
		fmt.Println("3. Consultar Minhas Caronas e Passageiros")
		fmt.Println("4. Cancelar Carona")
		fmt.Println("5. Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		switch opcao {
		case "1":
			autenticarMotorista(cliente, reader)
		case "2":
			publicarCarona(cliente, reader)
		case "3":
			consultarCaronas(cliente)
		case "4":
			cancelarCarona(cliente, reader)
		case "5":
			fmt.Println("Encerrando conexão do motorista...")
			return
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}

}
