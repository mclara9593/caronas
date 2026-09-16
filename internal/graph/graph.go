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
	mu    sync.RWMutex
	Nodes map[string][]Section // Origem -> Lista de trechos saindo da cidade
	Rides map[string]Ride      // IdRide -> Objeto da Carona
}

func NovoGrafo() *Grafo {
	return &Grafo{
		Nodes: make(map[string][]Section),
		Rides: make(map[string]Ride),
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
	defer g.mu.RUnlock()

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

			// Monta uma Ride combinada para o passageiro
			resultados = append(resultados, Ride{
				IdRide:     "search-result",
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

	return resultados
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

	// Aplicação da alteração nos trechos
	for _, id := range sectionIDs {
		for origem, list := range g.Nodes {
			for i, sec := range list {
				if sec.IdSection == id {
					g.Nodes[origem][i].Assentos += delta
				}
			}
		}
	}

	return true
}
