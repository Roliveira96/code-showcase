package main

import (
	"fmt"
	"log"

	"github.com/go-playground/locales/pt_BR"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	pt_translations "github.com/go-playground/validator/v10/translations/pt"
)

type Usuario struct {
	Nome     string `validate:"required,max=500"`
	Idade    int    `validate:"gte=0,lte=130"`
	CPF      string `validate:"required,len=11,numeric"`
	Email    string `validate:"required,email"`
	Telefone string `validate:"required,e164"`
}

func main() {
	validate, trans := configurarValidador()

	// Exemplos de usuários
	usuarios := []Usuario{
		{"João Silva", 30, "12345678901", "joao.silva@example.com", "+5511987654321"},      // Válido
		{"Maria Santos", 200, "98765432109", "maria.santos@example.com", "+5521912345678"}, // Idade inválida
		{"", -10, "12345", "invalido", "123456"},                                           // Totalmente inválido
	}

	for i, usuario := range usuarios {
		validarUsuario(validate, trans, i+1, usuario)
	}
}

func configurarValidador() (*validator.Validate, ut.Translator) {
	validate := validator.New()

	// Cria um tradutor universal
	pt := pt_BR.New()
	uni := ut.New(pt, pt)

	// Obtém o tradutor para português
	trans, found := uni.GetTranslator("pt_BR")
	if !found {
		log.Fatal("tradutor para pt_BR não encontrado")
	}

	// Registra o tradutor para o validador
	if err := pt_translations.RegisterDefaultTranslations(validate, trans); err != nil {
		log.Fatal(err)
	}

	// Traduções personalizadas (opcional)
	translations := map[string]string{
		"required": "{0} é obrigatório.",
		"max":      "{0} não pode ter mais de {1} caracteres.",
		"gte":      "{0} deve ser maior ou igual a {1}.",
		"lte":      "{0} deve ser menor ou igual a {1}.",
		"len":      "{0} deve ter exatamente {1} caracteres.",
		"email":    "{0} deve ser um email válido.",
		"e164":     "{0} deve estar no formato E.164.",
	}

	for tag, message := range translations {
		// Registra as traduções personalizadas
		if err := validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, message, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
			return t
		}); err != nil {
			log.Fatal(err)
		}
	}

	return validate, trans
}

func validarUsuario(validate *validator.Validate, trans ut.Translator, exemplo int, usuario Usuario) {
	fmt.Printf("\nExemplo %d:\n", exemplo)
	err := validate.Struct(usuario)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fmt.Printf("Erro no campo %s: %s\n", err.Field(), err.Translate(trans))
		}
	} else {
		fmt.Println("Validação bem-sucedida!")
	}
}
