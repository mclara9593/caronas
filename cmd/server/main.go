package main

import "github.com/mclara9593/caronas/internal/server"

func main() {
	srv := server.NovoServidor(":9593") // Porta do Servidor Central
	if err := srv.Iniciar(); err != nil {
		panic(err)
	}
}
