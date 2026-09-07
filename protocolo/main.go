package protocolo

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

type ConductorSingin struct {
	Login    []string `json:"login"`
	Password string   `json:"password"`
	CNH      int      `json:"cnh"`
}

type ConductorAuth struct {
	Login    []string `json:"login"`
	Password string   `json:"password"`
}

type UserAuth struct {
	Login    []string `json:"login"`
	Password string   `json:"password"`
}

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

//Confirmação de reserva por usuário
//Não precisa de payload

// Consulta de reserva por usuário
// Confirmação de reserva por usuário
type SearchBooking struct {
	RideID string `json:"ride_id"`
}

// Cancelar reserva por usuário
type CancelBooking struct {
	RideID string `json:"ride_id"`
}
