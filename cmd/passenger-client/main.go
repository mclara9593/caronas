package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

//IO

const ServerAddress = "localhost:12000"

func main() {
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		fmt.Printf("Erro ao conectar ao Servidor Central (%s): %v\n", ServerAddress, err)
		return
	}
	defer conn.Close()
	fmt.Println("Conectado ao Servidor Central com sucesso!")

	reader := bufio.NewReader(os.Stdin)

	//SYN

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
			autenticarPassageiro(conn, reader)
		case "2":
			buscarItinerarios(conn, reader)
		case "3":
			reservarItinerario(conn, reader)
		case "4":
			consultarReservas(conn)
		case "5":
			cancelarReserva(conn, reader)
		case "6":
			fmt.Println("Encerrando conexão do passageiro...")
			return
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}

func autenticarPassageiro(conn net.Conn, reader *bufio.Reader) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)
	fmt.Printf("Autenticando passageiro: %s...\n", user)
}

func buscarItinerarios(conn net.Conn, reader *bufio.Reader) {
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

func reservarItinerario(conn net.Conn, reader *bufio.Reader) {
	// A confirmação do itinerário (mesmo com vários trechos) deve ser atômica [7]
	fmt.Print("Digite o ID do itinerário/trechos que deseja reservar: ")
	itinerarioID, _ := reader.ReadString('\n')
	itinerarioID = strings.TrimSpace(itinerarioID)

	fmt.Printf("Iniciando transação atômica de reserva para o itinerário %s...\n", itinerarioID)
}

func consultarReservas(conn net.Conn) {
	fmt.Println("Buscando reservas ativas do usuário...")
}

func cancelarReserva(conn net.Conn, reader *bufio.Reader) {
	fmt.Print("Digite o ID da Reserva que deseja cancelar: ")
	reservaID, _ := reader.ReadString('\n')
	reservaID = strings.TrimSpace(reservaID)

	fmt.Printf("Solicitando cancelamento da reserva %s...\n", reservaID)
}
