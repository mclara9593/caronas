package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

// IP e porta do Servidor Central
const ServerAddress = "localhost:12000"

func main() {
	//cria o socket
	conn, err := net.Dial("tcp", ServerAddress)
	if err != nil {
		fmt.Printf("Erro ao conectar ao Servidor Central (%s): %v\n", ServerAddress, err)
		return
	}
	defer conn.Close()
	fmt.Println("Conectado ao Servidor Central com sucesso!")

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
			autenticarMotorista(conn, reader)
		case "2":
			publicarCarona(conn, reader)
		case "3":
			consultarCaronas(conn)
		case "4":
			cancelarCarona(conn, reader)
		case "5":
			fmt.Println("Encerrando conexão do motorista...")
			return
		default:
			fmt.Println("Opção inválida. Tente novamente.")
		}
	}
}

func autenticarMotorista(conn net.Conn, reader *bufio.Reader) {
	fmt.Print("Digite seu Usuário: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	// Aqui você monta a estrutura JSON da requisição e envia pelo socket [2]
	fmt.Printf("Enviando solicitação de autenticação para o usuário: %s...\n", user)
}

func publicarCarona(conn net.Conn, reader *bufio.Reader) {
	fmt.Print("Digite as cidades da rota separadas por vírgula (ex: Salvador,Feira,Conquista): ")
	rotaInput, _ := reader.ReadString('\n')
	rotaInput = strings.TrimSpace(rotaInput)

	fmt.Print("Data e Hora de Partida (ex: 2026-09-10 08:00): ")
	data, _ := reader.ReadString('\n')
	data = strings.TrimSpace(data)

	fmt.Print("Quantidade de assentos livres: ")
	assentos, _ := reader.ReadString('\n')
	assentos = strings.TrimSpace(assentos)

	fmt.Print("Preço cobrado por trecho (R$): ")
	preco, _ := reader.ReadString('\n')
	preco = strings.TrimSpace(preco)

	// Esses dados capturados serão serializados em JSON e enviados ao servidor [2]
	fmt.Println("Processando publicação de carona...")
}

func consultarCaronas(conn net.Conn) {
	fmt.Println("Requisitando histórico de caronas publicadas ao servidor...")
}

func cancelarCarona(conn net.Conn, reader *bufio.Reader) {
	fmt.Print("Digite o ID da Carona que deseja cancelar: ")
	id, _ := reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Printf("Enviando solicitação de cancelamento para a carona %s...\n", id)
}
