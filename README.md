<h3 align="center">VAIJUNTO</h3>

<p align="center">
  Sistema de caronas
</p>

<div align="center">

![Status](https://img.shields.io/badge/Status-Etapa%203%20-green)
![Language](https://img.shields.io/badge/Linguagem-GO-blue)

</div>

<div align="center">


[Sobre o projeto](#sobre-o-projeto) 
• [Requisitos](#requisitos) 
• [Instalação](#instalação) 
• [Solução](#solução)
• [Arquitetura](#arquitetura) 
• [ Estrutura e ferramentas](#estruturas-e-ferramentas-utilizadas) 
•  [Fluxo de Dados](#fluxo-de-dados) 
• [Testes e Resultados](#testes-e-resultados) 

</div>


## 📄Sobre o Projeto
Uma startup de mobilidade deseja lançar um serviço de caronas compartilhadas para viagens de média e
longa distância. Atualmente, a empresa mantém um servidor central, onde os motoristas publicam as
caronas que pretendem realizar e os passageiros consultam e reservam assentos pela Internet, sem
qualquer intermediação humana.


Além do servidor central que mantém o estado das caronas e das reservas, os seguintes componentes do
produto devem ser implementados e apresentados pelo aluno:




### Funcionais
| Requisito | Descrição |
| --- | --- |
| Autenticação de usuário | |
| Autenticação de motorista | |
| Publicação de carona | Publicação de carona contendo orota, data, assentos e preço por trecho  |
| Consulta de carona  | Motorista pode consultar carona já publicada bem como os passageiros |
| Cancelar carona | O motorista pode cancelar uma carona  |
| Buscar itnerários| O usuário buscar itinerários entre uma origem e um destino em uma data|
| Confirmação de reserva |O usuário pode confirmar a reserva de um itinerário composto por um ou mais trechos|
| Consulta de reserva  | Usuário pode consultar reserva |
| Cancelar reserva |  Usuário pode cancelar uma reserva  |



### Não Funcionais
| Requisito | Descrição |
| --- | --- |
| Interopabilidade |  |
| |  |


## 📦Execução

 

#### 1. Clonar o repositório
````
git clone [https://github.com/mclara9593/caronas) 

````


#### 2. Instalar o go
Instalar o pacote
````
rm -f go.linux-amd64.tar.gz go*.linux-amd64.tar.gz

export VERSAO_GO=$(curl -fsSL 'https://go.dev/VERSION?m=text' | head -n 1 | sed 's/^go//')

echo "Versão selecionada: $VERSAO_GO"

curl -fL -o "go${VERSAO_GO}.linux-amd64.tar.gz" \
  "https://go.dev/dl/go${VERSAO_GO}.linux-amd64.tar.gz"

file "go${VERSAO_GO}.linux-amd64.tar.gz"
````

Extrair
````
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf "go${VERSAO_GO}.linux-amd64.tar.gz"
````

Configurar o caminho
````
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
````


## 💻Solução

### Protocolo de comunicação cliente servidor

### Roteamento por grafos

#### Algoritmos de busca

## Domínio da solução
* Disponibilidade de assentos deve ser cobrado por trecho

* Passageiro que confirma primeiro recebe preferência

* Ao conectar o trecho A→B (Carona X) ao trecho B →C (Carona Y), a partida da segunda carona ocorre após a chegada da primeira:
Horario de partida B→C > Horario de partida A→B


## 👩🏻‍💻Arquitetura 

### 🖧 Componentes
  Socket: Interface de software pela qual um processo envia e recebe mensagens de e para a rede. Ele funciona de maneira lógica como a "porta" entre a camada de aplicação (controlada pelo desenvolvedor) e a camada de transporte (controlada pelo sistema operacional) dentro de um hospedeiro.




Esqueleto da requisição:
````
{
  "action": "NOME_DA_OPERACAO",
  "request_id": "ID_UNICO_DA_REQUISICAO",
  "payload": {
    "campo_especifico_1": "valor",
    "campo_especifico_2": "valor"
  }
}
````

## 🛠Estruturas e ferramentas utilizadas 
* Docker
* Go
* REST API
* P2P Gossip Protocol
* TCP/IP Protocol

## Referencias
https://www.geeksforgeeks.org/computer-networks/tcp-3-way-handshake-process/

## ✍️ Colaboradores

Este projeto foi desenvolvido por:

- [**Maria Clara**](https://github.com/)

