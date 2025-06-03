# Changelog

Todas as mudanças notáveis neste projeto serão documentadas neste arquivo.

O formato é baseado em [Keep a Changelog](https://keepachangelog.com/pt-BR/1.0.0/),
e este projeto adere ao [Semantic Versioning](https://semver.org/lang/pt-BR/).

## [Unreleased]

### Adicionado
- Documentação completa no README.md
- Arquivo CONTRIBUTING.md com diretrizes de contribuição
- Código de conduta (CODE_OF_CONDUCT.md)
- Templates de issue e pull request
- Pipeline CI/CD com GitHub Actions
- Configuração do golangci-lint
- Melhorias na validação de configuração
- Tratamento de erros aprimorado

### Alterado
- Limpeza de dependências desnecessárias no go.mod
- Melhoria na estrutura de erros customizados
- Otimização do gerenciamento de conexões

## [0.1.0] - 2024-01-XX

### Adicionado
- Implementação inicial da biblioteca multitenant
- Suporte para PostgreSQL como banco de dados principal
- Suporte para MongoDB como banco de dados alternativo
- Cache Redis com TTL configurável
- Middlewares HTTP para Gin, Fiber e Chi
- Gerenciamento de conexões com pool otimizado
- Context-aware tenant resolution
- CLI utilities para resolução de tenant
- Worker support para processamento em background
- Testes unitários e de integração abrangentes
- Configuração via variáveis de ambiente
- Clean Architecture com separação de responsabilidades

### Funcionalidades Core
- **TenantService**: Gerenciamento completo de tenants
- **ConnectionManager**: Pool de conexões por tenant e role
- **TenantCache**: Cache distribuído com Redis
- **TenantContext**: Propagação de tenant via context.Context
- **HTTP Middlewares**: Integração com frameworks populares
- **CLI Resolver**: Utilitários para aplicações não-HTTP

### Estrutura
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

### Configuração Suportada
- `MULTITENANT_DATABASE_TYPE`: postgres ou mongodb
- `MULTITENANT_DATABASE_DSN`: String de conexão do banco
- `MULTITENANT_REDIS_URL`: URL do Redis
- `MULTITENANT_CACHE_TTL`: TTL do cache
- `MULTITENANT_HEADER_NAME`: Nome do header HTTP
- `MULTITENANT_POOL_SIZE`: Tamanho do pool de conexões
- `MULTITENANT_MAX_RETRIES`: Máximo de tentativas
- `MULTITENANT_RETRY_DELAY`: Delay entre tentativas
- `MULTITENANT_LOG_LEVEL`: Nível de log
- `MULTITENANT_IGNORED_ENDPOINTS`: Lista de endpoints a serem ignorados pelo middleware, separados por vírgula (ex: /health,/metrics)

### Dependências Principais
- `github.com/gin-gonic/gin`: Framework HTTP
- `github.com/gofiber/fiber/v2`: Framework HTTP alternativo
- `github.com/go-chi/chi/v5`: Router HTTP minimalista
- `github.com/jackc/pgx/v5`: Driver PostgreSQL
- `go.mongodb.org/mongo-driver`: Driver MongoDB
- `github.com/redis/go-redis/v9`: Cliente Redis
- `github.com/google/uuid`: Geração de UUIDs
- `github.com/stretchr/testify`: Framework de testes

[Unreleased]: https://github.com/victorximenis/multitenant/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/victorximenis/multitenant/releases/tag/v0.1.0 