package protocol

import (
	"encoding/json"
)

type Request struct {
	Action    string          `json:"action"`     // Ex: "auth_conductor", "publish_ride"
	RequestID string          `json:"request_id"` // ID único gerado no cliente
	Payload   json.RawMessage `json:"payload"`    // Bytes brutos do JSON interno
}

type Response struct {
	RequestID string          `json:"request_id"` // Retorna o mesmo ID para o cliente se achar
	Status    string          `json:"status"`     // "SUCCESS" ou "ERROR"
	Message   string          `json:"message"`    // Descrição ou erro
	Payload   json.RawMessage `json:"payload"`    // Opcional: dados de retorno (ex: lista de caronas)
}

// ============================
// CADASTRO E LOGIN
// ============================

// CadastroPassageiro é o payload enviado para criar uma conta de passageiro
type CadastroPassageiro struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// CadastroMotorista é o payload enviado para criar uma conta de motorista.
// É igual ao de passageiro, acrescido apenas da CNH.
type CadastroMotorista struct {
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Senha string `json:"senha"`
	CNH   string `json:"cnh"`
}

// LoginRequest é usado tanto por passageiros quanto por motoristas para autenticar
type LoginRequest struct {
	Email string `json:"email"`
	Senha string `json:"senha"`
}

// LoginResult é devolvido pelo servidor dentro do Payload da Response quando o login dá certo,
// para que o cliente saiba o nome de quem acabou de logar
type LoginResult struct {
	Nome string `json:"nome"`
}

// ============================
// CARONAS / RESERVAS
// ============================

type PushRide struct {
	Route           []string `json:"route"`
	DepartureTime   string   `json:"departure_time"`
	Capacity        int      `json:"available_seats"`
	PricePerSegment float64  `json:"price_per_segment"`
}

// Cancelar e Consultar carona por motorista
type CancelSearchRide struct {
	Route  []string `json:"route"`
	RideID string   `json:"ride_id"`
}

// Buscar itinerários por usuário
type SearchRoute struct {
	Destination string `json:"destination"`
	Source      string `json:"source"`
	ArrivalTime string `json:"arrival_time"`
}

// Reserva, Consulta de reserva, e cancelamento por usuário
type GetID struct {
	RideID string `json:"ride_id"`
}
