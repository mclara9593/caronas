package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"

	protocol "github.com/mclara9593/caronas/internal/protocol"
)

// função para conectar ao servidor com tentativas limitadas
func conectarAoServidor(endereco string, maxTentativas int) (net.Conn, error) {
	var conn net.Conn
	var err error

	for i := 1; i <= maxTentativas; i++ {
		fmt.Printf(" Tentando conectar ao Servidor Central (%d/%d)...\n", i, maxTentativas)

		// net.DialTimeout evita que o cliente fique travado indefinidamente se a rede falhar
		conn, err = net.DialTimeout("tcp", endereco, 3*time.Second)
		if err == nil {
			return conn, nil // Conexão estabelecida com sucesso!
		}

		fmt.Printf("Falha na conexão: %v. Tentando novamente em 2 segundos...\n", err)
		time.Sleep(2 * time.Second)
	}

	return nil, err // Retorna o último erro se esgotar as tentativas
}

type Cliente struct {
	Nome          string
	Conn          net.Conn
	Reader        *bufio.Reader
	IsLogged      bool
	UsuarioLogado string
}

func newCliente(nome string, conn net.Conn) *Cliente {
	return &Cliente{
		Nome:     nome,
		Conn:     conn,
		Reader:   bufio.NewReader(conn),
		IsLogged: false,
	}
}

// EnviarRequisicao envia qualquer ação/payload e retorna a resposta do servidor
func (c *Cliente) EnviarRequisicao(action string, payload interface{}) (*protocol.Response, error) {
	// 1. Serializa o payload específico para bytes de JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar payload: %v", err)
	}

	// Monta o envelope genérico da requisição
	req := protocol.Request{
		Action:    action,
		RequestID: fmt.Sprintf("req-%d", time.Now().UnixNano()), // ID único por timestamp
		Payload:   payloadBytes,
	}

	// 3. Serializa o envelope completo para JSON
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar requisição: %v", err)
	}

	// 4. Envia pelo socket TCP com o delimitador '\n' para enquadramento [Kurose]
	_, err = c.Conn.Write(append(reqBytes, '\n'))
	if err != nil {
		return nil, fmt.Errorf("erro ao enviar dados pelo socket: %v", err)
	}

	// 5. Aguarda e lê a resposta vinda do servidor até o '\n'
	respostaBytes, err := c.Reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("erro ao ler resposta do servidor: %v", err)
	}

	// 6. Decodifica a resposta no struct Response
	var resp protocol.Response
	err = json.Unmarshal(respostaBytes, &resp)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta JSON: %v", err)
	}

	return &resp, nil
}
