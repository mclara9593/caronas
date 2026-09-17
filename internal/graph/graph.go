package graph

import (
	"fmt"
	"sync"
)

type Section struct {
	IdSection string  `json:"id"`
	Origem    string  `json:"origem"`
	Destino   string  `json:"destino"`
	Assentos  int     `json:"assentos"`
	Preco     float64 `json:"preco"`
}

// Estrutura que representa uma carona, que pode ser composta por múltiplos trechos
type Ride struct {
	IdRide     string    `json:"id"`
	Sections   []Section `json:"Sections"`
	IdDriver   string    `json:"id_driver"`
	TotalPrice float64   `json:"total_price"`
}

// Grafo representa a rede de viagens e trechos
type Grafo struct {
	mu          sync.RWMutex
	Nodes       map[string][]Section
	Rides       map[string]Ride
	Itineraries map[string][]string // id da rota -> lista de ids de trecho que ela engloba
	itinSeq     int
}

func NovoGrafo() *Grafo {
	return &Grafo{
		Nodes:       make(map[string][]Section),
		Rides:       make(map[string]Ride),
		Itineraries: make(map[string][]string),
	}
}

//==============================MANIPULAÇÃO DE CARONAS E TRECHOS======================================

// Quebra a rota do motorista em trechos e insere no Grafo
// AddRide agora recebe um preço por trecho (precos[i] corresponde ao trecho route[i] -> route[i+1]),
// em vez de um preço único aplicado a toda a rota.
func (g *Grafo) AddRide(idRide string, route []string, seats int, precos []float64, idDriver string) error {
	if len(precos) != len(route)-1 {
		return fmt.Errorf("esperava %d preço(s) para %d trecho(s), recebeu %d", len(route)-1, len(route)-1, len(precos))
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	var rideSections []Section
	var totalPrice float64

	for i := 0; i < len(route)-1; i++ {
		origem := route[i]
		destino := route[i+1]
		secID := idRide + "-sec-" + origem + "-" + destino
		preco := precos[i]

		sec := Section{
			IdSection: secID,
			Origem:    origem,
			Destino:   destino,
			Assentos:  seats,
			Preco:     preco,
		}

		rideSections = append(rideSections, sec)
		totalPrice += sec.Preco
		g.Nodes[origem] = append(g.Nodes[origem], sec)
	}

	g.Rides[idRide] = Ride{
		IdRide:     idRide,
		Sections:   rideSections,
		IdDriver:   idDriver,
		TotalPrice: totalPrice,
	}

	return nil
}

// SearchRoute faz a busca BFS e retorna as opções de viagens (Rides) encontradas

func (g *Grafo) SearchRoute(origem, destino string) []Ride {
	g.mu.RLock()

	var resultados []Ride

	type PathNode struct {
		CidadeAtual string
		Caminho     []Section
		Visitados   map[string]bool
	}

	fila := []PathNode{
		{
			CidadeAtual: origem,
			Caminho:     []Section{},
			Visitados:   map[string]bool{origem: true},
		},
	}

	for len(fila) > 0 {
		curr := fila[0]
		fila = fila[1:]

		if curr.CidadeAtual == destino {
			var total float64
			for _, sec := range curr.Caminho {
				total += sec.Preco
			}

			resultados = append(resultados, Ride{
				Sections:   curr.Caminho,
				IdDriver:   "multi",
				TotalPrice: total,
			})
			continue
		}

		for _, sec := range g.Nodes[curr.CidadeAtual] {
			if sec.Assentos > 0 && !curr.Visitados[sec.Destino] {
				novoVisitados := make(map[string]bool)
				for k, v := range curr.Visitados {
					novoVisitados[k] = v
				}
				novoVisitados[sec.Destino] = true

				novoCaminho := append([]Section{}, curr.Caminho...)
				novoCaminho = append(novoCaminho, sec)

				fila = append(fila, PathNode{
					CidadeAtual: sec.Destino,
					Caminho:     novoCaminho,
					Visitados:   novoVisitados,
				})
			}
		}
	}

	g.mu.RUnlock()

	// Registra cada itinerário encontrado com um ID único, guardando a lista
	// de trechos que ele engloba — assim o passageiro só precisa informar
	// esse ID e o servidor resolve todos os trechos por trás dele.
	g.mu.Lock()
	for i := range resultados {
		resultados[i].IdRide = g.registrarItinerario(resultados[i].Sections)
	}
	g.mu.Unlock()

	return resultados
}

// registrarItinerario cria um ID único para uma combinação de trechos (uma rota
// encontrada pela busca) e guarda o mapeamento id -> lista de ids de trecho.
// Pressupõe que o chamador já está segurando o Lock de escrita.
func (g *Grafo) registrarItinerario(sections []Section) string {
	g.itinSeq++
	id := fmt.Sprintf("rota-%d", g.itinSeq)

	ids := make([]string, len(sections))
	for i, sec := range sections {
		ids[i] = sec.IdSection
	}
	g.Itineraries[id] = ids

	return id
}

// ResolveItinerary traduz um ID de rota (gerado pela busca) na lista de ids de
// trecho que ela contém. Se o ID não for um itinerário registrado (ex: quando
// o motorista informa direto o id de um único trecho), assume que é o próprio
// id de um trecho isolado.
func (g *Grafo) ResolveItinerary(id string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if secoes, existe := g.Itineraries[id]; existe {
		return secoes
	}
	return []string{id}
}

// RidesByDriver retorna todas as caronas publicadas por um motorista específico (filtra pelo IdDriver)
func (g *Grafo) RidesByDriver(idDriver string) []Ride {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var resultado []Ride
	for _, ride := range g.Rides {
		if ride.IdDriver == idDriver {
			resultado = append(resultado, ride)
		}
	}
	return resultado
}

// RemoveRide cancela a carona informada e limpa as arestas do Grafo
func (g *Grafo) RemoveRide(idRide string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	ride, existe := g.Rides[idRide]
	if !existe {
		return
	}

	for _, secToRemove := range ride.Sections {
		origem := secToRemove.Origem
		sectionsOrigem := g.Nodes[origem]

		var atualizados []Section
		for _, sec := range sectionsOrigem {
			if sec.IdSection != secToRemove.IdSection {
				atualizados = append(atualizados, sec)
			}
		}

		if len(atualizados) == 0 {
			delete(g.Nodes, origem)
		} else {
			g.Nodes[origem] = atualizados
		}
	}

	delete(g.Rides, idRide)
}

// AdjustSeats decrementa ou incrementa assentos de forma atômica
// AdjustSeats decrementa ou incrementa assentos de forma atômica
func (g *Grafo) AdjustSeats(sectionIDs []string, delta int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validação prévia de capacidade
	if delta < 0 {
		for _, id := range sectionIDs {
			encontrado := false
			for _, list := range g.Nodes {
				for _, sec := range list {
					if sec.IdSection == id {
						encontrado = true
						if sec.Assentos+delta < 0 {
							return false
						}
					}
				}
			}
			if !encontrado {
				return false
			}
		}
	}

	// Aplicação da alteração nos trechos usados pela busca (g.Nodes)
	for _, id := range sectionIDs {
		for origem, list := range g.Nodes {
			for i, sec := range list {
				if sec.IdSection == id {
					g.Nodes[origem][i].Assentos += delta
				}
			}
		}
	}

	// Aplicação da alteração também na cópia guardada em g.Rides,
	// que é o que "Consultar Minhas Caronas" (RidesByDriver) exibe.
	// Sem isso, o motorista continua vendo o número antigo de assentos.
	for rideID, ride := range g.Rides {
		for i, sec := range ride.Sections {
			for _, id := range sectionIDs {
				if sec.IdSection == id {
					ride.Sections[i].Assentos += delta
				}
			}
		}
		g.Rides[rideID] = ride
	}

	return true
}
