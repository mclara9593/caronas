package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/mclara9593/caronas/internal/connection"
	"github.com/mclara9593/caronas/internal/protocol"
	//"github.com/mclara9593/caronas/cmd/client/auth"
)

func publicarCarona(cliente *connection.Cliente, reader *bufio.Reader) {
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

	// Coleta o preço por Section
	fmt.Print("Preço por Section (R$): ")
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

func consultarCaronas(cliente *connection.Cliente) {
	fmt.Println("Consultando caronas ativas do motorista...")
	// todo
}

func cancelarCarona(cliente **connection.Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID da carona que deseja cancelar: ")
	idCarona, _ := reader.ReadString('\n')
	idCarona = strings.TrimSpace(idCarona)

	// todo

	fmt.Printf("Cancelando carona com ID: %s...\n", idCarona)
}

func main() {
	//Tenta conectar ao Servidor Central
	conn, err := connection.("localhost:9593", 5)
	if err != nil {
		fmt.Printf(" Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	//Cria a instância do cliente
	cliente := 
	fmt.Println("Conectado com sucesso ao Servidor Central do VaiJunto!")

	// Executa o loop do menu
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- VAIJUNTO: PAINEL DO MOTORISTA ---")
		fmt.Println("1. Cadastrado (1º acesso) ")
		fmt.Println("2. Login")

		fmt.Println("1. Publicar Carona")
		fmt.Println("2. Consultar Minhas Caronas e Passageiros")
		fmt.Println("3. Cancelar Carona")
		fmt.Println("4. Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		//				if !cliente.IsLogged && (opcao == "2" || opcao == "3" || opcao == "4" || opcao == "5") {
		//			fmt.Println("\n Acesso negado! Você precisa fazer login (Opção 1) antes de realizar ações.")
		//			continue
		//		}

		switch opcao {
		case "1":
			connection.CadastrarMotorista(cliente, reader)
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
