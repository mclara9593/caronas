package domain

import "sync"

// Usuario representa os dados que guardamos de quem se cadastra no sistema.
// CNH fica vazia para passageiros e preenchida para motoristas.
type Usuario struct {
	Nome  string
	Email string
	Senha string
	CNH   string
}

// Repositorio guarda em memória os usuários cadastrados no sistema.
// O servidor mantém um Repositorio para passageiros e outro para motoristas,
// então a mesma pessoa pode ter uma conta de cada tipo com o mesmo email.
type Repositorio struct {
	mu       sync.RWMutex
	usuarios map[string]Usuario // Chave: Email do usuário
}

func NovoRepositorio() *Repositorio {
	return &Repositorio{
		usuarios: make(map[string]Usuario),
	}
}

// Cadastrar adiciona um novo usuário (se o email já não existir nesse repositório)
func (r *Repositorio) Cadastrar(u Usuario) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, existe := r.usuarios[u.Email]; existe {
		return false // já existe alguém com esse email
	}

	r.usuarios[u.Email] = u
	return true
}

// ValidarLogin checa se o email existe e a senha bate.
// Retorna os dados do usuário (para pegarmos o Nome, por exemplo) e um bool de sucesso.
func (r *Repositorio) ValidarLogin(email, senha string) (Usuario, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, existe := r.usuarios[email]
	if !existe || user.Senha != senha {
		return Usuario{}, false
	}

	return user, true
}

// Buscar retorna os dados de um usuário pelo email, sem checar senha.
// Útil, por exemplo, para associar uma reserva ao usuário logado.
func (r *Repositorio) Buscar(email string) (Usuario, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, existe := r.usuarios[email]
	return user, existe
}
