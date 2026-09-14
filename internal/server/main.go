package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/mclara9593/caronas/internal/graph"
	"github.com/mclara9593/caronas/internal/protocol"
)

// Server gerencia as conexões ativas e o estado compartilhado
type Server struct {
	grafo    *graph.Grafo
	addr     string
	listener net.Listener
	mu       sync.Mutex
}

// NovoServidor instancia o servidor TCP
func NovoServidor(addr string) *Server {
	return &Server{
		addr:  addr,
		grafo: graph.NovoGrafo(),
	}
}

// Iniciar abre a porta e escuta conexões TCP de clientes
func (s *Server) Iniciar() error {
	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("erro ao iniciar escuta TCP na porta %s: %w", s.addr, err)
	}
	s.listener = l
	defer s.listener.Close()

	fmt.Printf("Servidor Central na porta %s \n", s.addr)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Erro ao aceitar conexão: %v\n", err)
			continue
		}

		// Trata cada novo cliente em uma goroutine isolada
		go s.tratarCliente(conn)
	}
}

// tratarCliente gerencia o ciclo de vida do socket TCP do cliente
func (s *Server) tratarCliente(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Novo cliente conectado: %s\n", conn.RemoteAddr().String())

	reader := bufio.NewReader(conn)

	for {
		// Lê a linha JSON demarcada por '\n'
		linha, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Cliente %s desconectou.\n", conn.RemoteAddr().String())
			} else {
				fmt.Printf(" Erro de leitura do cliente %s: %v\n", conn.RemoteAddr().String(), err)
			}
			return
		}

		// Desempacota o envelope da requisição
		var req protocol.Request
		if err := json.Unmarshal(linha, &req); err != nil {
			s.enviarErro(conn, "", "Formato JSON inválido")
			continue
		}

		// Roteia a solicitação conforme a Ação (Action)
		s.rotearRequisicao(conn, &req)
	}
}

// rotearRequisicao lê o 'Action' e redireciona para a função correspondente
func (s *Server) rotearRequisicao(conn net.Conn, req *protocol.Request) {
	switch req.Action {

	// --- AÇÕES DO MOTORISTA ---
	case "auth_conductor":
		s.handleAuthConductor(conn, req)
	case "publish_ride":
		s.handlePublishRide(conn, req)
	case "cancel_ride":
		s.handleCancelRide(conn, req)

	// --- AÇÕES DO PASSAGEIRO ---
	case "auth_user":
		s.handleAuthUser(conn, req)
	case "search_route":
		s.handleSearchRoute(conn, req)
	case "book_ride":
		s.handleBookRide(conn, req)
	case "get_bookings":
		s.handleGetBookings(conn, req)
	case "cancel_booking":
		s.handleCancelBooking(conn, req)

	default:
		s.enviarErro(conn, req.RequestID, "Ação não reconhecida pelo servidor")
	}
}

// ==========================================
// HANDLERS DO MOTORISTA
// ==========================================

func (s *Server) handleAuthConductor(conn net.Conn, req *protocol.Request) {
	var payload protocol.ConductorAuth
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de autenticação inválido")
		return
	}

	// TODO: Validar credenciais na memória/storage
	s.enviarSucesso(conn, req.RequestID, "Motorista autenticado com sucesso!", nil)
}

// PUBLICAR CARONA
func (s *Server) handlePublishRide(conn net.Conn, req *protocol.Request) {
	var payload protocol.PushRide
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de publicação inválido")
		return
	}

	// Gera um ID único para a carona/ride
	rideID := "ride-" + req.RequestID

	// Chama o Grafo passando os dados extraídos do payload
	s.grafo.AddRide(
		rideID,
		payload.Route,
		payload.Capacity,
		payload.PricePerSegment,
		"driver-default",
	)

	s.enviarSucesso(conn, req.RequestID, "Carona publicada e inserida no Grafo!", nil)
}

func (s *Server) handleCancelRide(conn net.Conn, req *protocol.Request) {
	var payload protocol.GetID
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload inválido")
		return
	}

	// Remove a carona e desfaz as arestas no Grafo
	s.grafo.RemoveRide(payload.RideID)
	s.enviarSucesso(conn, req.RequestID, "Carona removida do Grafo com sucesso!", nil)
}

// ==========================================
// HANDLERS DO PASSAGEIRO
// ==========================================

func (s *Server) handleAuthUser(conn net.Conn, req *protocol.Request) {
	var payload protocol.UserAuth
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload inválido")
		return
	}

	// TODO: Validar dados do cliente no Storage em memória
	s.enviarSucesso(conn, req.RequestID, "Passageiro autenticado com sucesso!", nil)
}

// Buscar rota de viagem
func (s *Server) handleSearchRoute(conn net.Conn, req *protocol.Request) {
	var payload protocol.SearchRoute
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de busca inválido")
		return
	}

	// Executa a busca em largura no Grafo
	resultados := s.grafo.SearchRoute(payload.Source, payload.Destination)

	// Converte os resultados para JSON antes de devolver ao cliente
	data, _ := json.Marshal(resultados)
	s.enviarSucesso(conn, req.RequestID, "Itinerários encontrados", data)
}

// Efetuar reserva
func (s *Server) handleBookRide(conn net.Conn, req *protocol.Request) {
	var payload protocol.GetID
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de reserva inválido")
		return
	}

	// Tenta decrementar 1 assento de forma atômica
	// O RideID aqui pode ser o ID de um trecho ou uma lista de trechos do itinerário
	sucesso := s.grafo.AdjustSeats([]string{payload.RideID}, -1)
	if !sucesso {
		s.enviarErro(conn, req.RequestID, "Não há assentos disponíveis para esta reserva.")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Reserva confirmada com sucesso!", nil)
}

func (s *Server) handleGetBookings(conn net.Conn, req *protocol.Request) {
	// TODO: Retornar histórico do passageiro
	s.enviarSucesso(conn, req.RequestID, "Reservas encontradas", nil)
}

func (s *Server) handleCancelBooking(conn net.Conn, req *protocol.Request) {
	var payload protocol.GetID
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de cancelamento inválido")
		return
	}

	// Incrementa a vaga de volta ao Grafo (+1 assento)
	s.grafo.AdjustSeats([]string{payload.RideID}, 1)
	s.enviarSucesso(conn, req.RequestID, "Reserva cancelada e vaga liberada!", nil)
}

// AUXILIARES DE RESPOSTA TCP/JSON

func (s *Server) enviarSucesso(conn net.Conn, reqID string, msg string, payload interface{}) {
	var rawPayload json.RawMessage
	if payload != nil {
		rawPayload, _ = json.Marshal(payload)
	}

	resp := protocol.Response{
		RequestID: reqID,
		Status:    "SUCCESS",
		Message:   msg,
		Payload:   rawPayload,
	}

	s.escreverResponse(conn, resp)
}

func (s *Server) enviarErro(conn net.Conn, reqID string, msg string) {
	resp := protocol.Response{
		RequestID: reqID,
		Status:    "ERROR",
		Message:   msg,
	}

	s.escreverResponse(conn, resp)
}

func (s *Server) escreverResponse(conn net.Conn, resp protocol.Response) {
	bytes, _ := json.Marshal(resp)
	conn.Write(append(bytes, '\n'))
}
