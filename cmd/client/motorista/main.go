package motorista

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
