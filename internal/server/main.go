package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/mclara9593/caronas/internal/domain"
	"github.com/mclara9593/caronas/internal/graph"
	"github.com/mclara9593/caronas/internal/protocol"
)

// Server gerencia as conexões ativas e o estado compartilhado
type Server struct {
	grafo           *graph.Grafo
	repoPassageiros *domain.Repositorio
	repoMotoristas  *domain.Repositorio
	addr            string
	listener        net.Listener
	mu              sync.Mutex
}

// NovoServidor instancia o servidor TCP
func NovoServidor(addr string) *Server {
	return &Server{
		addr:            addr,
		grafo:           graph.NovoGrafo(),
		repoPassageiros: domain.NovoRepositorio(),
		repoMotoristas:  domain.NovoRepositorio(),
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

	// --- CADASTRO E LOGIN ---
	case "cadastrar_user":
		s.handleCadastrarUser(conn, req)
	case "cadastrar_motorista":
		s.handleCadastrarMotorista(conn, req)
	case "auth_user":
		s.handleAuthUser(conn, req)
	case "auth_conductor":
		s.handleAuthConductor(conn, req)

	// --- AÇÕES DO MOTORISTA ---
	case "publish_ride":
		s.handlePublishRide(conn, req)
	case "cancel_ride":
		s.handleCancelRide(conn, req)
	case "get_my_rides":
		s.handleGetMyRides(conn, req)

	// --- AÇÕES DO PASSAGEIRO ---
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

// ==========================
// CADASTRO E LOGIN
// ==========================

// Cadastro de passageiro
func (s *Server) handleCadastrarUser(conn net.Conn, req *protocol.Request) {
	var payload protocol.CadastroPassageiro
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de cadastro inválido")
		return
	}

	if payload.Nome == "" || payload.Email == "" || payload.Senha == "" {
		s.enviarErro(conn, req.RequestID, "Nome, email e senha são obrigatórios")
		return
	}

	usuario := domain.Usuario{
		Nome:  payload.Nome,
		Email: payload.Email,
		Senha: payload.Senha,
	}

	if !s.repoPassageiros.Cadastrar(usuario) {
		s.enviarErro(conn, req.RequestID, "Já existe um passageiro cadastrado com este email")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Passageiro cadastrado com sucesso! Faça login para continuar.", nil)
}

// Cadastro de motorista (mesmos dados do passageiro + CNH)
func (s *Server) handleCadastrarMotorista(conn net.Conn, req *protocol.Request) {
	var payload protocol.CadastroMotorista
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de cadastro inválido")
		return
	}

	if payload.Nome == "" || payload.Email == "" || payload.Senha == "" || payload.CNH == "" {
		s.enviarErro(conn, req.RequestID, "Nome, email, CNH e senha são obrigatórios")
		return
	}

	usuario := domain.Usuario{
		Nome:  payload.Nome,
		Email: payload.Email,
		Senha: payload.Senha,
		CNH:   payload.CNH,
	}

	if !s.repoMotoristas.Cadastrar(usuario) {
		s.enviarErro(conn, req.RequestID, "Já existe um motorista cadastrado com este email")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Motorista cadastrado com sucesso! Faça login para continuar.", nil)
}

// Login de passageiro
func (s *Server) handleAuthUser(conn net.Conn, req *protocol.Request) {
	var payload protocol.LoginRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de login inválido")
		return
	}

	usuario, ok := s.repoPassageiros.ValidarLogin(payload.Email, payload.Senha)
	if !ok {
		s.enviarErro(conn, req.RequestID, "Email ou senha inválidos")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Passageiro autenticado com sucesso!", protocol.LoginResult{Nome: usuario.Nome})
}

// Login de motorista
func (s *Server) handleAuthConductor(conn net.Conn, req *protocol.Request) {
	var payload protocol.LoginRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de login inválido")
		return
	}

	usuario, ok := s.repoMotoristas.ValidarLogin(payload.Email, payload.Senha)
	if !ok {
		s.enviarErro(conn, req.RequestID, "Email ou senha inválidos")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Motorista autenticado com sucesso!", protocol.LoginResult{Nome: usuario.Nome})
}

// ==========================
// HANDLERS DO MOTORISTA
// ===========================

// PUBLICAR CARONA
func (s *Server) handlePublishRide(conn net.Conn, req *protocol.Request) {
	var payload protocol.PushRide
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de publicação inválido")
		return
	}

	if len(payload.Route) < 2 {
		s.enviarErro(conn, req.RequestID, "A rota precisa ter pelo menos 2 cidades")
		return
	}

	if len(payload.Precos) != len(payload.Route)-1 {
		s.enviarErro(conn, req.RequestID, fmt.Sprintf(
			"Rota com %d trecho(s) precisa de %d preço(s), mas veio %d",
			len(payload.Route)-1, len(payload.Route)-1, len(payload.Precos),
		))
		return
	}

	// Gera um ID único para a carona/ride
	rideID := "ride-" + req.RequestID

	// Chama o Grafo passando os dados extraídos do payload
	if err := s.grafo.AddRide(
		rideID,
		payload.Route,
		payload.Capacity,
		payload.Precos,
		payload.DriverEmail,
	); err != nil {
		s.enviarErro(conn, req.RequestID, err.Error())
		return
	}

	// Devolve o ID gerado para o motorista guardar (usado depois em cancel_ride)
	s.enviarSucesso(conn, req.RequestID, "Carona publicada e inserida no Grafo!", protocol.PublishRideResult{
		RideID: rideID,
	})
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

// Lista as caronas publicadas por um motorista específico
func (s *Server) handleGetMyRides(conn net.Conn, req *protocol.Request) {
	var payload protocol.GetMyRidesRequest
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload inválido")
		return
	}

	if payload.Email == "" {
		s.enviarErro(conn, req.RequestID, "Email do motorista é obrigatório")
		return
	}

	rides := s.grafo.RidesByDriver(payload.Email)

	msg := fmt.Sprintf("%d carona(s) encontrada(s)", len(rides))
	s.enviarSucesso(conn, req.RequestID, msg, rides)
}

// ==========================================
// HANDLERS DO PASSAGEIRO
// ==========================================

// Buscar rota de viagem
func (s *Server) handleSearchRoute(conn net.Conn, req *protocol.Request) {
	var payload protocol.SearchRoute
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de busca inválido")
		return
	}

	// Executa a busca em largura no Grafo
	resultados := s.grafo.SearchRoute(payload.Source, payload.Destination)

	// Passa a struct direto — enviarSucesso já faz o Marshal internamente
	s.enviarSucesso(conn, req.RequestID, "Itinerários encontrados", resultados)
}

// Efetuar reserva
func (s *Server) handleBookRide(conn net.Conn, req *protocol.Request) {
	var payload protocol.GetID
	if err := json.Unmarshal(req.Payload, &payload); err != nil {
		s.enviarErro(conn, req.RequestID, "Payload de reserva inválido")
		return
	}

	// payload.RideID pode ser o id de uma rota inteira (com vários trechos)
	// ou o id de um único trecho — ResolveItinerary trata os dois casos.
	sectionIDs := s.grafo.ResolveItinerary(payload.RideID)

	sucesso := s.grafo.AdjustSeats(sectionIDs, -1)
	if !sucesso {
		s.enviarErro(conn, req.RequestID, "Não há assentos disponíveis para esta reserva.")
		return
	}

	s.enviarSucesso(conn, req.RequestID, "Reserva confirmada com sucesso!", protocol.BookRideResult{
		RideID: payload.RideID,
	})
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

	sectionIDs := s.grafo.ResolveItinerary(payload.RideID)
	s.grafo.AdjustSeats(sectionIDs, 1)
	s.enviarSucesso(conn, req.RequestID, "Reserva cancelada e vaga liberada!", nil)
}

// ==================================
// AUXILIARES DE RESPOSTA TCP/JSON
// ==================================

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
