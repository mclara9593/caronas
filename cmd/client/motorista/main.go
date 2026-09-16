package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mclara9593/caronas/cmd/client/auth"
	"github.com/mclara9593/caronas/internal/connection"
	"github.com/mclara9593/caronas/internal/protocol"
)

func publicarCarona(cliente *connection.Cliente, reader *bufio.Reader) {
	// 1. Coleta a rota completa (ex: Feira de Santana, Salvador, Aracaju),
	// repetindo a pergunta até o motorista informar pelo menos 2 cidades.
	var rota []string
	for {
		fmt.Print("Digite as cidades da rota (separadas por vírgula): ")
		rotaInput, _ := reader.ReadString('\n')
		rotaInput = strings.TrimSpace(rotaInput)

		cidadesRaw := strings.Split(rotaInput, ",")
		rota = nil
		for _, c := range cidadesRaw {
			cidadeTratada := strings.TrimSpace(c)
			if cidadeTratada != "" {
				rota = append(rota, cidadeTratada)
			}
		}

		if len(rota) < 2 {
			fmt.Println(" A rota precisa ter pelo menos 2 cidades (origem e destino).")
			continue
		}
		break
	}

	// Quebra a rota em trechos de 2 em 2 (origem -> destino) e pede um preço
	// individual para cada um, em vez de um preço único pra rota inteira.
	precos := make([]float64, len(rota)-1)
	fmt.Println("\nAgora informe o preço de cada trecho:")
	for i := 0; i < len(rota)-1; i++ {
		origem := rota[i]
		destino := rota[i+1]

		for {
			fmt.Printf("Trecho %d: %s → %s | Preço (R$): ", i+1, origem, destino)
			precoInput, _ := reader.ReadString('\n')
			precoInput = strings.TrimSpace(precoInput)
			precoInput = strings.Replace(precoInput, ",", ".", 1) // aceita vírgula como separador decimal

			valor, err := strconv.ParseFloat(precoInput, 64)
			if err != nil || valor <= 0 {
				fmt.Println(" Digite um valor numérico maior que zero (ex: 25.50).")
				continue
			}
			precos[i] = valor
			break
		}
	}

	// Coleta a data/horário de partida
	fmt.Print("Horário de partida (ex: 2026-09-12 18:00): ")
	depTime, _ := reader.ReadString('\n')
	depTime = strings.TrimSpace(depTime)

	// Coleta a capacidade de assentos (mesmo reader usado acima, sem misturar com Scanln)
	var vagas int
	for {
		fmt.Print("Quantidade de vagas disponíveis: ")
		vagasInput, _ := reader.ReadString('\n')
		vagasInput = strings.TrimSpace(vagasInput)

		valor, err := strconv.Atoi(vagasInput)
		if err != nil || valor <= 0 {
			fmt.Println(" Digite um número inteiro maior que zero.")
			continue
		}
		vagas = valor
		break
	}

	// Monta o payload conforme a struct PushRide.
	// DriverEmail é o email de quem está logado — é isso que permite depois
	// filtrar "minhas caronas" corretamente.
	payload := protocol.PushRide{
		Route:         rota,
		DepartureTime: depTime,
		Capacity:      vagas,
		Precos:        precos,
		DriverEmail:   cliente.UsuarioLogado,
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
	// Endereço do servidor agora é configurável via linha de comando.
	// Se não for informado, usa localhost:9593 como padrão.
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
