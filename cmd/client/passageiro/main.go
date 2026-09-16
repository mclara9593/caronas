package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mclara9593/caronas/cmd/client/auth"
	"github.com/mclara9593/caronas/internal/connection"
	"github.com/mclara9593/caronas/internal/protocol"
)

func buscarItinerarios(cliente *connection.Cliente, reader *bufio.Reader) {
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

	resp, err := cliente.EnviarRequisicao("search_route", payload)
	if err != nil {
		fmt.Printf("❌ Erro de comunicação: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
	if len(resp.Payload) > 0 {
		fmt.Println(string(resp.Payload))
	}
}

func reservarItinerario(cliente *connection.Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID do itinerário/ride para reservar: ")
	idRide, _ := reader.ReadString('\n')
	idRide = strings.TrimSpace(idRide)

	payload := protocol.GetID{RideID: idRide}

	resp, err := cliente.EnviarRequisicao("book_ride", payload)
	if err != nil {
		fmt.Printf("❌ Erro de comunicação: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func consultarReservas(cliente *connection.Cliente) {
	resp, err := cliente.EnviarRequisicao("get_bookings", protocol.GetID{})
	if err != nil {
		fmt.Printf("❌ Erro de comunicação: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func cancelarReserva(cliente *connection.Cliente, reader *bufio.Reader) {
	fmt.Print("ID da Reserva: ")
	idReserva, _ := reader.ReadString('\n')
	idReserva = strings.TrimSpace(idReserva)

	payload := protocol.GetID{RideID: idReserva}

	resp, err := cliente.EnviarRequisicao("cancel_booking", payload)
	if err != nil {
		fmt.Printf("❌ Erro de comunicação: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func main() {
	conn, err := connection.ConectarAoServidor("localhost:9593", 5)
	if err != nil {
		fmt.Printf("Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	cliente := connection.NewCliente("Passageiro", conn)
	fmt.Println("✔ Conectado com sucesso ao Servidor Central do VaiJunto!")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- VAIJUNTO: PAINEL DO PASSAGEIRO ---")
		fmt.Println("1. Cadastrar (1º acesso)")
		fmt.Println("2. Login")
		fmt.Println("3. Buscar Itinerários de Carona")
		fmt.Println("4. Reservar Itinerário")
		fmt.Println("5. Consultar Minhas Reservas")
		fmt.Println("6. Cancelar Reserva")
		fmt.Println("7. Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		// Barreira de Autenticação: opções 3 a 6 exigem login
		precisaLogin := opcao == "3" || opcao == "4" || opcao == "5" || opcao == "6"
		if precisaLogin && !cliente.IsLogged {
			fmt.Println("\n🔒 Acesso negado! Você precisa fazer login (Opção 2) antes de realizar ações.")
			continue
		}

		switch opcao {
		case "1":
			if cliente.IsLogged {
				fmt.Printf("Você já está logado como %s!\n", cliente.NomeLogado)
				continue
			}
			auth.CadastrarPassageiro(cliente, reader)
		case "2":
			if cliente.IsLogged {
				fmt.Printf("Você já está logado como %s!\n", cliente.NomeLogado)
				continue
			}
			auth.LoginPassageiro(cliente, reader)
		case "3":
			buscarItinerarios(cliente, reader)
		case "4":
			reservarItinerario(cliente, reader)
		case "5":
			consultarReservas(cliente)
		case "6":
			cancelarReserva(cliente, reader)
		case "7":
			fmt.Println("Encerrando conexão...")
			return
		default:
			fmt.Println("Opção inválida.")
		}
	}
}
