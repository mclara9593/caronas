package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mclara9593/caronas/internal/protocol"
)

func conectarAoServidor(endereco string, maxTentativas int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 1; i <= maxTentativas; i++ {
		fmt.Printf(" Tentando conectar ao Servidor Central (%d/%d)...\n", i, maxTentativas)

		conn, err = net.DialTimeout("tcp", endereco, 3*time.Second)
		if err == nil {
			return conn, nil
		}

		fmt.Printf(" Falha na conexão: %v. Tentando novamente em 2 segundos...\n", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err
}

type Cliente struct {
	Nome   string
	Conn   net.Conn
	Reader *bufio.Reader
}

func newCliente(nome string, conn net.Conn) *Cliente {
	return &Cliente{
		Nome:   nome,
		Conn:   conn,
		Reader: bufio.NewReader(conn),
	}
}

func (c *Cliente) EnviarRequisicao(action string, payload interface{}) (*protocol.Response, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %v", err)
	}

	req := protocol.Request{
		Action:    action,
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()),
		Payload:   payloadBytes,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	_, err = c.Conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar dados pelo socket: %v", err)
	}

	respostaBytes, err := c.Reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta do servidor: %v", err)
	}

	var resp protocol.Response
	err = json.Unmarshal(respostaBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	return &resp, nil
}

func autenticarPassageiro(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Print("Digite sua Senha: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	authPayload := protocol.UserAuth{
		Login:    []string{user},
		Password: password,
	}

	resp, err := cliente.EnviarRequisicao("auth_user", authPayload)
	if err != nil {
		fmt.Printf("❌ Erro ao enviar requisição: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

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

	searchPayload := protocol.SearchRoute{
		Source:      origem,
		Destination: destino,
		ArrivalTime: data,
	}

	resp, err := cliente.EnviarRequisicao("search_route", searchPayload)
	if err != nil {
		fmt.Printf("❌ Erro na busca: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func reservarItinerario(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID do itinerário/trechos que deseja reservar: ")
	itinerarioID, _ := reader.ReadString('\n')
	itinerarioID = strings.TrimSpace(itinerarioID)

	bookPayload := protocol.SearchBooking{
		RideID: itinerarioID,
	}

	resp, err := cliente.EnviarRequisicao("book_ride", bookPayload)
	if err != nil {
		fmt.Printf("❌ Erro ao reservar: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func consultarReservas(cliente *Cliente) {
	resp, err := cliente.EnviarRequisicao("get_bookings", nil)
	if err != nil {
		fmt.Printf("❌ Erro ao consultar: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func cancelarReserva(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID da Reserva que deseja cancelar: ")
	reservaID, _ := reader.ReadString('\n')
	reservaID = strings.TrimSpace(reservaID)

	cancelPayload := protocol.CancelBooking{
		RideID: reservaID,
	}

	resp, err := cliente.EnviarRequisicao("cancel_booking", cancelPayload)
	if err != nil {
		fmt.Printf("❌ Erro ao cancelar: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func main() {
	conn, err := conectarAoServidor("localhost: 9593", 5)
	if err != nil {
		fmt.Printf("❌ Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	cliente := newCliente("Passageiro", conn)
	fmt.Println("✔ Conectado com sucesso ao Servidor Central do VaiJunto!")

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
