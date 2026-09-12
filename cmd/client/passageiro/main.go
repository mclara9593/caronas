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

func autenticarPassageiro(cliente *Cliente, reader *bufio.Reader) (string, interface{}, bool) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Print("Digite sua Senha: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	payload := protocol.UserAuth{
		Login:    []string{user},
		Password: password,
	}

	return "auth_user", payload, true
}

func buscarItinerarios(cliente *Cliente, reader *bufio.Reader) (string, interface{}, bool) {
	fmt.Print("Cidade de Origem: ")
	origem, _ := reader.ReadString('\n')
	origem = strings.TrimSpace(origem)

	fmt.Print("Cidade de Destino: ")
	destino, _ := reader.ReadString('\n')
	destino = strings.TrimSpace(destino)

	fmt.Print("Data desejada (ex: 2026-09-10): ")
	data, _ := reader.ReadString('\n')
	data = strings.TrimSpace(data)

	payload := protocol.SearchRoute{
		Source:      origem,
		Destination: destino,
		ArrivalTime: data,
	}

	return "search_route", payload, true
}

func reservarItinerario(cliente *Cliente, reader *bufio.Reader) {
	fmt.Print("Cidade de Origem: ")
	origem, _ := reader.ReadString('\n')
	origem = strings.TrimSpace(origem)

	fmt.Print("Cidade de Destino: ")
	destino, _ := reader.ReadString('\n')
	destino = strings.TrimSpace(destino)

	fmt.Print("Data desejada (ex: 2026-09-10): ")
	data, _ := reader.ReadString('\n')
	data = strings.TrimSpace(data)

	payload := protocol.SearchRoute{
		Source:      origem,
		Destination: destino,
		ArrivalTime: data,
	}

	return "book_ride", payload, true

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

		var action string
		var payload interface{}
		var deveEnviar bool

		switch opcao {
		case "1":
			action, payload, deveEnviar = autenticarPassageiro(cliente, reader)
		case "2":
			action, payload, deveEnviar = buscarItinerarios(cliente, reader)
		case "3":
			action, payload, deveEnviar = reservarItinerario(cliente, reader)
		case "4":
			action, payload, deveEnviar = consultarReservas(cliente)
		case "5":
			action, payload, deveEnviar = cancelarReserva(cliente, reader)
		case "6":
			fmt.Println("Encerrando conexão...")
			return
		default:
			fmt.Println("Opção inválida.")
			continue
		}

		//Envia a requisição se a função capturou os dados com sucesso
		if deveEnviar {
			resp, err := cliente.EnviarRequisicao(action, payload)
			if err != nil {
				fmt.Printf("❌ Erro de comunicação: %v\n", err)
			} else {
				fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
			}
		}
	}
}
