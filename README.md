# 🚀 Code Showcase - Ricardo de Oliveira

Bem-vindo ao meu **Code Showcase**! Este repositório é um compilado de experimentos, testes de conceito (PoC), protótipos e ferramentas úteis desenvolvidas em **Go (Golang)**.

Aqui concentro meus estudos e implementações de referência para resolver problemas reais de engenharia de software, desde automações de infraestrutura até integrações complexas de API.

## 👨‍💻 Sobre Mim

**Ricardo de Oliveira**
*Backend Developer | Golang Enthusiast*

Sou apaixonado por criar soluções eficientes e escaláveis. Este portfólio demonstra minha experiência prática com diversas tecnologias e padrões de arquitetura.

*   🌐 **Site**: [studio4you.com.br](https://studio4you.com.br)
*   📱 **WhatsApp**: [Entrar em contato](https://wa.me/5542999486362)

## 🛠️ Tecnologias & Tópicos

Este repositório cobre uma vasta gama de tópicos, incluindo:

*   **Autenticação & Segurança**: OAuth2, PKCE, Azure AD, Keycloak.
*   **Integrações**: APIs REST, Email (Outlook/SMTP), Kafka, MinIO, Ollama (IA).
*   **Processamento de Arquivos**: Excel, geração de imagens, OCR (Tesseract).
*   **DevOps & Observabilidade**: OpenTelemetry (OTEL), Monitoramento de Desktop, Pprof.
*   **Automação**: Web Scraping, Geradores (Domínios, Favicons).

## 📂 Catálogo de Projetos

Abaixo você encontra a lista de projetos organizados por categoria:

### 🔐 Autenticação e Segurança
| Projeto | Descrição |
| :--- | :--- |
| `example-auth-azure-ad-go` | Integração e autenticação usando Azure AD. |
| `example-authorization-code-grant-with-pkce` | Implementação do fluxo Authorization Code com PKCE. |
| `example-automatize-add-roles-for-keycloak` | Automação para adicionar roles/funções no Keycloak. |
| `example-crypto-code` | Exemplos e testes de criptografia. |

### ☁️ Integrações, APIs e Web
| Projeto | Descrição |
| :--- | :--- |
| `example-api-users` | API REST básica para gerenciamento de usuários. |
| `example-api-use-ollama` | Integração com LLMs locais usando Ollama. |
| `example-minio-site-builder` | Ferramenta para construção de sites estáticos com MinIO. |
| `example-web-scraping-golang` | Scrapers para extração de dados da web. |
| `example-test-response-kafka` | Testes de produção/consumo de mensagens com Kafka. |

### 📧 Notificações e Email
| Projeto | Descrição |
| :--- | :--- |
| `example-notification-go` | Sistema genérico de notificações. |
| `example-send-email-use-apikey-azure-outlook` | Envio de emails via API do Outlook/Azure. |
| `example-send-emails-user-api` | API dedicada a envio de emails transacionais. |
| `example-generate-json-from-send-email` | Utilitário para gerar JSON a partir de logs de envio. |
| `example-send-loop-notification-test` | Testes de carga e loop de notificações. |

### 📊 Dados, Excel e Imagens
| Projeto | Descrição |
| :--- | :--- |
| `example-process-excel-file` | Leitura e processamento de arquivos Excel. |
| `example-export-template-excel` | Geração e exportação de templates em Excel. |
| `example-export-image-grafic-radar` | Geração de gráficos de radar exportados como imagem. |
| `example-get-text-by-tesseract` | Extração de texto de imagens usando OCR (Tesseract). |
| `example-favicon-generator` | Gerador automático de favicons. |

### 💻 Sistema e Monitoramento
| Projeto | Descrição |
| :--- | :--- |
| `example-app-windows` | Desenvolvimento de aplicações nativas para Windows. |
| `example-monitor-desktop` | Monitoramento de recursos do sistema desktop. |
| `example-otel-collactor` | Configuração e uso do OpenTelemetry Collector. |
| `example-use-pprof-tracker` | Profiling e análise de performance com Pprof. |
| `example-all-ips` | Utilitário para listar todos os IPs da rede. |
| `example-execution-list` | Gerenciamento de filas/listas de execução local. |

## 🚀 Como Executar

Cada pasta é um módulo ou projeto independente. Para rodar qualquer um deles, navegue até a pasta e execute:

```bash
cd nome-do-exemplo
go mod tidy
go run main.go
```

*Nota: Verifique se existem arquivos `.env.example` que precisam ser renomeados para `.env` com suas configurações locais.*

---
© 2026 Ricardo de Oliveira. All rights reserved.
