package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	// O servidor escuta passivamente na porta 12000
	listener, err := net.Listen("tcp", ":12000")
	if err != nil {
		fmt.Printf("Erro ao iniciar o servidor: %v\n", err)
		return
	}
	defer listener.Close()
	fmt.Println("Servidor Central escutando na porta 12000...")

	for {
		// Aceita conexões de entrada de forma concorrente
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %v\n", err)
			continue
		}
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Cliente conectado: %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)
	for {
		messageBytes, err := reader.ReadBytes('\n')
		if err != nil {
			return // Cliente desconectou
		}
		fmt.Printf("Recebido do cliente: %s", string(messageBytes))
		conn.Write([]byte("Mensagem processada pelo servidor!\n"))
	}
}
