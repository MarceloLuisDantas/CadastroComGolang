package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// Pessoa representa uma pessoa
type Pessoa struct {
	Nome  string `json:"nome"`
	Idade int    `json:"ida"`
	Cpf   string `json:"cpf"`
}

// Busca retorna uma lista de Pessoas pelos dados do banco
func Busca(db *sql.DB, q string) (map[int]Pessoa, error) {

	// Cria um Map para as pessoas cadastras, onde a chave é o ID
	pessoas := make(map[int]Pessoa)

	// Realiza a query e verifica caso tenho ocorrio um erro, caso tenha ocorrido o erro é retornado
	result, err := db.Query(q)
	if err != nil {
		return pessoas, err
	}
	// Garante que "result" sera fechado ao final da função
	defer result.Close()

	// Percorre os valores do resultado e adicionar na lista
	for result.Next() {
		var id int
		p := Pessoa{}
		// Scan pega os valores das coluna, e passa para um valor,
		// No caso é prenchido os dados da pessoa junto ao seu id
		err = result.Scan(&p.Nome, &p.Idade, &p.Cpf, &id)
		if err != nil {
			return pessoas, err
		}
		pessoas[id] = p
	}

	// Fecha "result" e verifica caso tenho ocorrido um erro
	err = result.Close()
	if err != nil {
		return pessoas, err
	}
	return pessoas, nil
}

// Cadastrar faz o cadastro de uma pessoa no banco de dados
func Cadastrar(db *sql.DB, p Pessoa) error {
	// Prepara um INSERT de a ser executado para cadastrar o usuario
	stmt, err := db.Prepare("INSERT INTO Pessoas (nome, idade, cpf) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}

	// Garante que "stmt" sera fechado
	defer stmt.Close()

	// Executa a query antes preparada
	_, err = stmt.Exec(p.Nome, p.Idade, p.Cpf)
	if err != nil {
		return err
	}
	return nil
}

// GetNome pega o nome do usuario
func GetNome(input *bufio.Scanner) string {
	fmt.Print("Digite seu nome: ")
	input.Scan()
	nome := input.Text()
	if nome == "" {
		fmt.Println("Nome invalido. tente novamente")
		return GetNome(input)
	}
	return nome
}

// GetIdade pega a idade do usuario
func GetIdade(input *bufio.Scanner) int {
	fmt.Print("Digite sua idade: ")
	input.Scan()
	idade, err := strconv.ParseInt(input.Text(), 10, 64)
	if err != nil {
		fmt.Println("Erro ao processar idade. tente novamente")
		return GetIdade(input)
	}
	return int(idade)
}

// FormatCPF Formata o CPF para formato aceitado
func FormatCPF(c string) string {
	nuns := "01213456789"

	// New String
	var NS []string

	// Remove todos os simbulos não numeros
	for _, v := range c {
		if strings.Contains(nuns, string(v)) {
			NS = append(NS, string(v))
		}
	}

	// Concatena a lista de Strings em uma String unica
	cpf := strings.Join(NS, "")
	return cpf
}

// CPFExiste verifica se o CPF já foi cadastrado
func CPFExiste(db *sql.DB, cpf string) bool {
	// Resgata do banco alguma linha que possua o CPF, caso n tenho um erro é retorando.
	var i int
	err := db.QueryRow("SELECT id FROM Pessoas WHERE cpf = ?", cpf).Scan(&i)
	if err != nil {
		return false
	}
	return true
}

// GetCpf pega o CPF do usuario
func GetCpf(db *sql.DB, input *bufio.Scanner) string {
	fmt.Print("Digite o seu CPF sem pontos ou traços: ")
	input.Scan()

	// Formata o CPF dado pelo usuario para remover os simbulos n numeros
	cpf := FormatCPF(input.Text())

	// Verifica se o CPF já foi cadastrado
	if !CPFExiste(db, cpf) {
		return cpf
	}
	fmt.Println("CPF já cadastro, tente novamente")
	return GetCpf(db, input)
}

// GeraPessoa gera uma pessoa
func GeraPessoa(db *sql.DB) Pessoa {
	input := bufio.NewScanner(os.Stdin)

	nome := GetNome(input)
	idade := GetIdade(input)
	cpf := GetCpf(db, input)
	pessoa := Pessoa{nome, idade, cpf}

	return pessoa
}

func main() {
	// Cria a conexão com o banco de dados
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/Pessoas")
	if err != nil {
		log.Fatalln("Não foi possivel acessar o banco de dados:", err)
	}
	defer db.Close()

	// Cria pessoa
	pessoa := GeraPessoa(db)

	// Cadastra a pessoa no banco
	err = Cadastrar(db, pessoa)
	if err != nil {
		log.Fatalln("Erro durante o cadastro:", err)
		// me, _ := err.(*mysql.MySQLError)

		// fmt.Println(pessoa.Cpf)
		// if int(me.Number) == 1062 {
		// 	fmt.Printf("O CPF %s já foi casdastrado, tente novamente\n", pessoa.Cpf)
		// }
	}
	fmt.Printf("%s cadastrada(o) com sucesso\n", pessoa.Nome)

	// Realiza uma busca no banco de dados
	var query string = "SELECT nome, idade, cpf, id FROM Pessoas ORDER BY `id`"
	pessoas, err := Busca(db, query)
	if err != nil {
		log.Fatalln("Erro durante operação:", err)
	}
	if len(pessoas) > 0 {
		for i, p := range pessoas {
			fmt.Printf("%d = %s. %d anos. CPF %s \n", i, p.Nome, p.Idade, p.Cpf)
		}
	} else {
		fmt.Println("Não a cadastros")
	}
}
