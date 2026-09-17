package graph_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mclara9593/caronas/internal/graph"
)

// TestConcorrenciaAdjustSeats simula 20 clientes tentando reservar assentos
// no mesmo trecho ao mesmo tempo, verificando que o número de reservas
// bem-sucedidas é exatamente igual à capacidade publicada (nem mais, nem
// menos) e que nenhuma corrida de dados permite "overbooking".
func TestConcorrenciaAdjustSeats(t *testing.T) {
	g := graph.NovoGrafo()

	const capacidade = 10
	const numClientes = 20

	err := g.AddRide(
		"ride-teste",
		[]string{"FSA", "SSA"},
		capacidade,
		[]float64{50.0},
		"motorista-teste@exemplo.com",
	)
	if err != nil {
		t.Fatalf("erro ao publicar carona de teste: %v", err)
	}

	// ID do único trecho gerado por essa carona (FSA -> SSA)
	sectionID := "ride-teste-sec-FSA-SSA"

	var wg sync.WaitGroup
	var sucessos int64
	var falhas int64

	for i := 0; i < numClientes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok := g.AdjustSeats([]string{sectionID}, -1)
			if ok {
				atomic.AddInt64(&sucessos, 1)
			} else {
				atomic.AddInt64(&falhas, 1)
			}
		}()
	}

	wg.Wait()

	if sucessos != capacidade {
		t.Errorf("esperava exatamente %d reservas bem-sucedidas, obteve %d", capacidade, sucessos)
	}

	esperadoFalhas := int64(numClientes - capacidade)
	if falhas != esperadoFalhas {
		t.Errorf("esperava %d falhas, obteve %d", esperadoFalhas, falhas)
	}

	t.Logf("Resultado: %d reservas confirmadas, %d recusadas por falta de vaga (capacidade era %d)", sucessos, falhas, capacidade)
}

// TestConcorrenciaReservaItinerarioCompleto simula vários passageiros buscando
// e reservando o mesmo itinerário multi-trecho ao mesmo tempo, usando o fluxo
// real do servidor: SearchRoute -> ResolveItinerary -> AdjustSeats.
func TestConcorrenciaReservaItinerarioCompleto(t *testing.T) {
	g := graph.NovoGrafo()

	const capacidade = 5
	const numClientes = 25

	err := g.AddRide(
		"ride-trecho1",
		[]string{"FSA", "SSA"},
		capacidade,
		[]float64{50.0},
		"motorista1@exemplo.com",
	)
	if err != nil {
		t.Fatalf("erro ao publicar 1º trecho: %v", err)
	}

	err = g.AddRide(
		"ride-trecho2",
		[]string{"SSA", "ARA"},
		capacidade,
		[]float64{70.0},
		"motorista2@exemplo.com",
	)
	if err != nil {
		t.Fatalf("erro ao publicar 2º trecho: %v", err)
	}

	itinerarios := g.SearchRoute("FSA", "ARA")
	if len(itinerarios) == 0 {
		t.Fatal("nenhum itinerário encontrado entre FSA e ARA")
	}

	idRota := itinerarios[0].IdRide

	var wg sync.WaitGroup
	var sucessos int64
	var falhas int64

	for i := 0; i < numClientes; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			sectionIDs := g.ResolveItinerary(idRota)
			ok := g.AdjustSeats(sectionIDs, -1)
			if ok {
				atomic.AddInt64(&sucessos, 1)
			} else {
				atomic.AddInt64(&falhas, 1)
			}
		}()
	}

	wg.Wait()

	if sucessos != capacidade {
		t.Errorf("esperava exatamente %d reservas bem-sucedidas no itinerário completo, obteve %d", capacidade, sucessos)
	}

	t.Logf("Itinerário %s: %d reservas confirmadas, %d recusadas (capacidade era %d por trecho)", idRota, sucessos, falhas, capacidade)
}
