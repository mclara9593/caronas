package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mclara9593/caronas/cmd/client/auth"
	"github.com/mclara9593/caronas/internal/connection"
	"github.com/mclara9593/caronas/internal/protocol"
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

	// Monta o payload conforme a struct PushRide.
	// DriverEmail é o email de quem está logado — é isso que permite depois
	// filtrar "minhas caronas" corretamente.
	payload := protocol.PushRide{
		Route:           rota,
		DepartureTime:   depTime,
		Capacity:        vagas,
		PricePerSegment: preco,
		DriverEmail:     cliente.UsuarioLogado,
	}

	// Envia a requisição via socket
	resp, err := cliente.EnviarRequisicao("publish_ride", payload)
	if err != nil {
		fmt.Printf("Erro ao publicar carona: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)

	// Se deu certo, o servidor manda o ID da carona dentro do Payload.
	// Repara no "null": se o payload vier vazio, o JSON manda literalmente a palavra
	// null (4 bytes) em vez de nada — por isso não basta checar len(resp.Payload) > 0.
	if resp.Status == "SUCCESS" && len(resp.Payload) > 0 && string(resp.Payload) != "null" {
		var resultado protocol.PublishRideResult
		if err := json.Unmarshal(resp.Payload, &resultado); err == nil && resultado.RideID != "" {
			fmt.Printf("🚗 ID da carona: %s (guarde esse ID para cancelar depois)\n", resultado.RideID)
		}
	}
}

func consultarCaronas(cliente *connection.Cliente) {
	resp, err := cliente.EnviarRequisicao("get_my_rides", protocol.GetMyRidesRequest{
		Email: cliente.UsuarioLogado,
	})
	if err != nil {
		fmt.Printf("Erro ao consultar caronas: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)

	if resp.Status != "SUCCESS" || len(resp.Payload) == 0 {
		return
	}

	var caronas []protocol.RideInfo
	if err := json.Unmarshal(resp.Payload, &caronas); err != nil {
		fmt.Printf("Erro ao interpretar a resposta: %v\n", err)
		return
	}

	if len(caronas) == 0 {
		fmt.Println("Você ainda não publicou nenhuma carona.")
		return
	}

	for _, carona := range caronas {
		fmt.Printf("\n🚗 Carona %s — total R$ %.2f\n", carona.IdRide, carona.TotalPrice)
		for _, trecho := range carona.Sections {
			fmt.Printf("   %s → %s | %d vaga(s) | R$ %.2f | id: %s\n",
				trecho.Origem, trecho.Destino, trecho.Assentos, trecho.Preco, trecho.IdSection)
		}
	}
}

func cancelarCarona(cliente *connection.Cliente, reader *bufio.Reader) {
	fmt.Print("Digite o ID da carona que deseja cancelar: ")
	idCarona, _ := reader.ReadString('\n')
	idCarona = strings.TrimSpace(idCarona)

	resp, err := cliente.EnviarRequisicao("cancel_ride", protocol.GetID{RideID: idCarona})
	if err != nil {
		fmt.Printf("Erro ao cancelar carona: %v\n", err)
		return
	}

	fmt.Printf("📩 Resposta do Servidor: [%s] %s\n", resp.Status, resp.Message)
}

func main() {
	// Tenta conectar ao Servidor Central
	conn, err := connection.ConectarAoServidor("localhost:9593", 5)
	if err != nil {
		fmt.Printf(" Erro ao conectar: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Cria a instância do cliente
	cliente := connection.NewCliente("Motorista", conn)
	fmt.Println("Conectado com sucesso ao Servidor Central do VaiJunto!")

	// Executa o loop do menu
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n--- VAIJUNTO: PAINEL DO MOTORISTA ---")
		fmt.Println("1. Cadastrar (1º acesso)")
		fmt.Println("2. Login")
		fmt.Println("3. Publicar Carona")
		fmt.Println("4. Consultar Minhas Caronas e Passageiros")
		fmt.Println("5. Cancelar Carona")
		fmt.Println("6. Sair")
		fmt.Print("Escolha uma opção: ")

		opcao, _ := reader.ReadString('\n')
		opcao = strings.TrimSpace(opcao)

		// Barreira de Autenticação: opções 3 a 5 exigem login
		precisaLogin := opcao == "3" || opcao == "4" || opcao == "5"
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
			auth.CadastrarMotorista(cliente, reader)
		case "2":
			if cliente.IsLogged {
				fmt.Printf("Você já está logado como %s!\n", cliente.NomeLogado)
				continue
			}
			auth.LoginMotorista(cliente, reader)
		case "3":
			publicarCarona(cliente, reader)
		case "4":
			consultarCaronas(cliente)
		case "5":
			cancelarCarona(cliente, reader)
		case "6":
			fmt.Println("Encerrando conexão do motorista...")
			return
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}
