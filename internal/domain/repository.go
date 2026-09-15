package domain

import "sync"

// Representa os dados que queremos guardar do cliente
type Usuario struct {
	Login string
	Senha string
}

// Repositorio guarda todos os usuários cadastrados no sistema
type Repositorio struct {
	mu       sync.RWMutex
	usuarios map[string]Usuario // Chave: Login do usuário
}

func NovoRepositorio() *Repositorio {
	return &Repositorio{
		usuarios: make(map[string]Usuario),
	}
}

// Cadastrar adiciona um novo usuário (se já não existir)
func (r *Repositorio) Cadastrar(login, senha string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.usuarios[login]; existe {
		return false // Usuário já existe
	}

	r.usuarios[login] = Usuario{Login: login, Senha: senha}
	return true
}

// ValidarLogin checa se o usuário existe e a senha bate
func (r *Repositorio) ValidarLogin(login, senha string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, existe := r.usuarios[login]
	if !existe {
		return false
	}

	return user.Senha == senha
}
