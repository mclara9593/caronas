package main

import (
	"bufio"
	"encoding/json"
	"flag"
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

	if resp.Status != "SUCCESS" || len(resp.Payload) == 0 {
		return
	}

	var itinerarios []protocol.RideInfo
	if err := json.Unmarshal(resp.Payload, &itinerarios); err != nil {
		fmt.Printf("Erro ao interpretar a resposta: %v\n", err)
		return
	}

	if len(itinerarios) == 0 {
		fmt.Println("Nenhum itinerário encontrado para essa rota.")
		return
	}

	for i, itin := range itinerarios {
		fmt.Printf("\nItinerário %d:\n", i+1)
		fmt.Printf("  ID da rota: %s\n", itin.IdRide)
		for _, trecho := range itin.Sections {
			fmt.Printf("  Saída: %s\n", trecho.Origem)
			fmt.Printf("  Destino: %s\n", trecho.Destino)
			fmt.Printf("  Quantidade de assentos: %d\n", trecho.Assentos)
		}
		fmt.Printf("  Valor total da carona: R$ %.2f\n", itin.TotalPrice)
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

	// Se deu certo, o servidor manda o ID da reserva dentro do Payload.
	// (o "null" != len(...) é pra evitar imprimir ID vazio se o payload vier nulo)
	if resp.Status == "SUCCESS" && len(resp.Payload) > 0 && string(resp.Payload) != "null" {
		var resultado protocol.BookRideResult
		if err := json.Unmarshal(resp.Payload, &resultado); err == nil && resultado.RideID != "" {
			fmt.Printf("🎫 Reserva confirmada! ID: %s (guarde esse ID para cancelar depois)\n", resultado.RideID)
		}
	}
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
	endereco := flag.String("addr", "localhost:9593", "Endereço do Servidor Central (ip:porta)")
	flag.Parse()

	conn, err := connection.ConectarAoServidor(*endereco, 5)
	if err != nil {
		fmt.Printf(" Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	cliente := connection.NewCliente("Motorista", conn)
	fmt.Printf("Conectado com sucesso ao Servidor Central do VaiJunto em %s!\n", *endereco)

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
