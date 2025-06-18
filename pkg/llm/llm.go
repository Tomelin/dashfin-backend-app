package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	genaiModel        = "gemini-2.5-pro"
	promptToParseNFCE = `
Título do Prompt: Extração Estruturada de Dados de Nota Fiscal Eletrônica (NF-e)

Instruções:

Acesse a URL fornecida de uma Nota Fiscal Eletrônica (NF-e).

Extração:

Faça o parse do conteúdo da NF-e para identificar e extrair todos os itens detalhados na nota.

Dados a serem extraídos por item:

Descrição do item: Nome completo ou detalhe do produto/serviço.
Valor total do item: O valor total daquele item específico (quantidade * valor unitário).
Dados adicionais a serem extraídos:

CNPJ do estabelecimento emissor da NF-e.
Formato de Saída:

Retorne os dados em formato JSON, estruturado da seguinte forma:

JSON

{
  "cnpj_estabelecimento": "12.345.678/0001-90",
  ""itens": [
    {
      "descricao_item": "Produto A",
      "valor_total_item": 150.75
    },
    {
      "descricao_item": "Serviço B",
      "valor_total_item": 200.00
    },
    {
      "descricao_item": "Produto C",
      "valor_total_item": 50.25
    }
  ]
}
`
)

type AgentInterface interface {
	Run(ctx context.Context, query string) ([]byte, error)
}

// Agent representa um indivíduo agente de IA com sua própria sessão e role
type Agent struct {
	Session *genai.ChatSession
}

func NewAgent() (AgentInterface, error) {
	apikey := os.Getenv("GEMINI_API_KEY")
	if apikey == "" {
		return nil, errors.New("GEMINI_API_KEY not set")
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apikey))
	if err != nil {
		return nil, fmt.Errorf("error creating client: %w", err)
	}

	model := client.GenerativeModel(genaiModel)
	// Configurações de segurança básicas replicadas do código original
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockNone},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockNone},
	}

	model.SetTemperature(0.0)
	model.SetCandidateCount(int32(100))
	model.SystemInstruction = genai.NewUserContent(genai.Text(promptToParseNFCE))
	model.ResponseMIMEType = "application/json"

	// CORREÇÃO: Inicia uma nova sessão de chat SEM argumentos
	cs := model.StartChat()

	// CORREÇÃO: Define o histórico inicial (incluindo a role) atribuindo-o diretamente à propriedade History
	cs.History = []*genai.Content{
		// A primeira entrada simula o "usuário" (orquestrador/sistema) dando a instrução de role
		{
			Parts: []genai.Part{genai.Text("I'm need that you access the URL send by user and execute the query to get a itens, price of itens and CNPJ of the seller")},
			Role:  "user",
		},
		// Uma resposta simulada do modelo para confirmar que entendeu a role
		{
			Parts: []genai.Part{genai.Text("Entendido. Estou pronto para processar as NFCe e extrair os dados solicitados no formato JSON.")},
			Role:  "model", // CORREÇÃO: Esta entrada deve ser do role "model"
		},
	}

	agent := &Agent{
		Session: cs,
	}
	return agent, nil
}

func (a *Agent) Run(ctx context.Context, query string) ([]byte, error) {
	message := fmt.Sprintf(`user: I'm need that you access this %s URL and folowing these steps:
	1. Get the body of URL
	2. Parse the body to get a CNPJ of the seller
	3. Parse the body to get a itens, price of itens
	4. Format the response to JSON
	5. No create random message
	6. Return the JSON`, query)

	log.Println(message)
	a.Session.History = append(a.Session.History, &genai.Content{
		Parts: []genai.Part{genai.Text(message)},
		Role:  "user",
	})
	resp, err := a.Session.SendMessage(ctx, genai.Text(message))
	if err != nil {
		return nil, fmt.Errorf("error sending message: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("no response from model")
	}

	log.Println(resp.Candidates[0].Content.Parts)

	response := fmt.Sprintf("%s", resp.Candidates[0].Content.Parts[0])

	b, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("error marshaling response: %w", err)
	}

	return b, nil
}
