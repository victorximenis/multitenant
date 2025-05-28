# Contribuindo para Multitenant Go Library

Obrigado por considerar contribuir para o projeto! Este documento fornece diretrizes e informações sobre como contribuir efetivamente.

## 📋 Índice

- [Código de Conduta](#código-de-conduta)
- [Como Contribuir](#como-contribuir)
- [Configuração do Ambiente](#configuração-do-ambiente)
- [Diretrizes de Desenvolvimento](#diretrizes-de-desenvolvimento)
- [Processo de Pull Request](#processo-de-pull-request)
- [Reportando Bugs](#reportando-bugs)
- [Sugerindo Melhorias](#sugerindo-melhorias)
- [Documentação](#documentação)

## 📜 Código de Conduta

Este projeto adere ao [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). Ao participar, você deve seguir este código. Por favor, reporte comportamentos inaceitáveis para [victorximenis@gmail.com].

## 🤝 Como Contribuir

Existem várias maneiras de contribuir:

- 🐛 **Reportar bugs**
- 💡 **Sugerir melhorias**
- 📝 **Melhorar documentação**
- 🔧 **Implementar features**
- 🧪 **Escrever testes**
- 📖 **Criar exemplos**

## 🛠️ Configuração do Ambiente

### Pré-requisitos

- Go 1.21 ou superior
- Docker e Docker Compose (para testes de integração)
- Git

### Setup Local

1. **Fork e clone o repositório**:
```bash
git clone https://github.com/SEU_USERNAME/multitenant.git
cd multitenant
```

2. **Instalar dependências**:
```bash
go mod download
```

3. **Configurar ambiente de teste**:
```bash
# Copiar arquivo de configuração de teste
cp testconfig.env.example testconfig.env

# Editar configurações se necessário
vim testconfig.env
```

4. **Iniciar dependências para testes**:
```bash
# PostgreSQL
docker run -d --name postgres-test -p 5432:5432 \
  -e POSTGRES_USER=dev_user \
  -e POSTGRES_PASSWORD=dev_password \
  -e POSTGRES_DB=multitenant_db \
  postgres:15

# Redis
docker run -d --name redis-test -p 6379:6379 redis:7

# MongoDB
docker run -d --name mongo-test -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=password \
  mongo:7
```

5. **Executar testes**:
```bash
go test ./...
```

## 📏 Diretrizes de Desenvolvimento

### Estrutura do Código

```
multitenant/
├── core/                 # Entidades e interfaces de domínio
├── infra/               # Implementações de infraestrutura
│   ├── connection/      # Gerenciamento de conexões
│   ├── mongodb/         # Repository MongoDB
│   ├── postgres/        # Repository PostgreSQL
│   └── redis/           # Cache Redis
├── interfaces/          # Camada de interface
│   ├── cli/            # Utilitários CLI
│   └── http/           # Middlewares HTTP
├── tenantcontext/       # Utilitários de contexto
└── examples/           # Exemplos de uso
```

### Padrões de Código

1. **Siga as convenções Go**:
   - Use `gofmt` para formatação
   - Execute `go vet` para análise estática
   - Use `golangci-lint` para linting

2. **Nomenclatura**:
   - Interfaces terminam com sufixo (ex: `TenantRepository`)
   - Métodos públicos começam com letra maiúscula
   - Use nomes descritivos e claros

3. **Documentação**:
   - Todas as funções públicas devem ter comentários Go doc
   - Comentários devem começar com o nome da função
   - Inclua exemplos quando apropriado

4. **Tratamento de Erros**:
   - Sempre trate erros explicitamente
   - Use erros customizados quando apropriado
   - Inclua contexto suficiente nas mensagens de erro

### Testes

1. **Cobertura**:
   - Mantenha cobertura de testes acima de 80%
   - Teste casos de sucesso e falha
   - Inclua testes de concorrência quando relevante

2. **Tipos de Teste**:
   - **Unitários**: Testam componentes isolados
   - **Integração**: Testam interação entre componentes
   - **End-to-end**: Testam fluxos completos

3. **Convenções**:
   - Arquivos de teste terminam com `_test.go`
   - Use `testify` para assertions
   - Organize testes em subtests quando apropriado

```go
func TestTenantService_GetTenant(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *core.Tenant
        wantErr bool
    }{
        {
            name:  "valid tenant",
            input: "test-tenant",
            want:  &core.Tenant{Name: "test-tenant"},
        },
        {
            name:    "tenant not found",
            input:   "nonexistent",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Commits

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: adiciona suporte para MySQL
fix: corrige memory leak no connection manager
docs: atualiza README com exemplos de MongoDB
test: adiciona testes para cache Redis
refactor: melhora estrutura do TenantService
```

Tipos de commit:
- `feat`: Nova funcionalidade
- `fix`: Correção de bug
- `docs`: Documentação
- `test`: Testes
- `refactor`: Refatoração
- `perf`: Melhoria de performance
- `chore`: Tarefas de manutenção

## 🔄 Processo de Pull Request

1. **Antes de começar**:
   - Verifique se já existe uma issue relacionada
   - Discuta mudanças grandes antes de implementar
   - Fork o repositório

2. **Durante o desenvolvimento**:
   - Crie uma branch descritiva: `feat/mysql-support`
   - Faça commits pequenos e focados
   - Mantenha a branch atualizada com `main`

3. **Antes de submeter**:
   - Execute todos os testes: `go test ./...`
   - Execute linting: `golangci-lint run`
   - Atualize documentação se necessário
   - Adicione entrada no CHANGELOG.md

4. **Submissão**:
   - Crie PR com título descritivo
   - Preencha template do PR
   - Marque reviewers apropriados
   - Responda a feedback construtivamente

### Template de Pull Request

```markdown
## Descrição
Breve descrição das mudanças.

## Tipo de Mudança
- [ ] Bug fix
- [ ] Nova feature
- [ ] Breaking change
- [ ] Documentação

## Checklist
- [ ] Testes passando
- [ ] Documentação atualizada
- [ ] CHANGELOG.md atualizado
- [ ] Linting sem erros

## Testes
Descreva os testes realizados.

## Screenshots (se aplicável)
```

## 🐛 Reportando Bugs

Use o [template de issue para bugs](.github/ISSUE_TEMPLATE/bug_report.md):

1. **Título claro**: Descreva o problema em uma linha
2. **Reprodução**: Passos para reproduzir o bug
3. **Comportamento esperado**: O que deveria acontecer
4. **Comportamento atual**: O que está acontecendo
5. **Ambiente**: Versão do Go, OS, versão da biblioteca
6. **Logs**: Inclua logs relevantes

## 💡 Sugerindo Melhorias

Use o [template de issue para features](.github/ISSUE_TEMPLATE/feature_request.md):

1. **Problema**: Que problema a feature resolve?
2. **Solução**: Descreva a solução proposta
3. **Alternativas**: Outras soluções consideradas
4. **Contexto**: Informações adicionais

## 📖 Documentação

### Tipos de Documentação

1. **README.md**: Visão geral e quick start
2. **Go doc**: Documentação de API
3. **Exemplos**: Código de exemplo funcional
4. **Wiki**: Guias detalhados e tutoriais

### Diretrizes

- Use linguagem clara e concisa
- Inclua exemplos de código
- Mantenha atualizada com mudanças
- Teste exemplos de código

## 🏷️ Versionamento

Seguimos [Semantic Versioning](https://semver.org/):

- **MAJOR**: Mudanças incompatíveis na API
- **MINOR**: Funcionalidades compatíveis
- **PATCH**: Correções compatíveis

## 🎯 Áreas Prioritárias

Contribuições são especialmente bem-vindas em:

- 📊 **Observabilidade**: Métricas, logging, tracing
- 🔒 **Segurança**: Validação, sanitização, auditoria
- 🚀 **Performance**: Otimizações, benchmarks
- 📚 **Documentação**: Guias, tutoriais, exemplos
- 🧪 **Testes**: Cobertura, casos edge, performance

## 📞 Contato

- **Issues**: [GitHub Issues](https://github.com/victorximenis/multitenant/issues)
- **Discussões**: [GitHub Discussions](https://github.com/victorximenis/multitenant/discussions)
- **Email**: victorximenis@gmail.com

## 🙏 Reconhecimento

Todos os contribuidores são reconhecidos no arquivo [CONTRIBUTORS.md](CONTRIBUTORS.md).

---

**Obrigado por contribuir! 🎉** 