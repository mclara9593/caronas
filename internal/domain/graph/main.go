package main
import (
	"sync"
	"time"
)

type Section struct {
	IdSection	   	string  `json:"id"`
	Origem   string  `json:"origem"`
	Destino  string  `json:"destino"`
	Assentos int     `json:"assentos"`
	Preco    float64 `json:"preco"`
}


// Estrutura que representa uma carona, que pode ser composta por múltiplos trechos
type Ride struct {
	IdRide		string   `json:"id"`
	Sections  []Section `json:"Sections"`
	IdDriver		string   `json:"id_driver"`
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
func (g *Grafo) AddRide(idRide string, route []string, seats int, price float64, idDriver string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var rideSections []Section
	var totalPrice float64

	for i := 0; i < len(route)-1; i++ {
		origem := route[i]
		destino := route[i+1]
		secID := idRide + "-sec-" + origem + "-" + destino

		sec := Section{
			ID:       secID,
			Origem:   origem,
			Destino:  destino,
			Assentos: seats,
			Preco:    price,
		}

		rideSections = append(rideSections, sec)
		totalPrice += price
		g.Nodes[origem] = append(g.Nodes[origem], sec)
	}

	g.Rides[idRide] = Ride{
		IdRide:     idRide,
		Sections:   rideSections,
		IdDriver:   idDriver,
		TotalPrice: totalPrice,
	}
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
			if sec.ID != secToRemove.ID {
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
					if sec.ID == id {
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
				if sec.ID == id {
					g.Nodes[origem][i].Assentos += delta
				}
			}
		}
	}

	return true
}