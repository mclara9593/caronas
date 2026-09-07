package main

import "fmt"

// Mapa representa um Mapa usando lista de adjacência
type Mapa struct {
	adjacencias map[string][]string
}

// NovoMapa cria um novo Mapa
func NovoMapa() *Mapa {
	return &Mapa{
		adjacencias: make(map[string][]string),
	}
}

// AdicionarVertice adiciona um vértice ao Mapa
func (g *Mapa) AdicionarVertice(vertice string) {
	if _, ok := g.adjacencias[vertice]; !ok {
		g.adjacencias[vertice] = []string{}
	}
}

// AdicionarAresta adiciona uma aresta (não direcionada) entre dois vértices
func (g *Mapa) AdicionarAresta(origem, destino string) {
	g.AdicionarVertice(origem)
	g.AdicionarVertice(destino)

	g.adjacencias[origem] = append(g.adjacencias[origem], destino)
	g.adjacencias[destino] = append(g.adjacencias[destino], origem)
}

// Imprimir exibe a lista de adjacência
func (g *Mapa) Imprimir() {
	for vertice, vizinhos := range g.adjacencias {
		fmt.Printf("%s -> %v\n", vertice, vizinhos)
	}
}

func main() {
	g := NovoMapa()
	g.AdicionarAresta("A", "B")
	g.AdicionarAresta("A", "C")
	g.AdicionarAresta("B", "D")

	g.Imprimir()
}
