# Multitenant Go Library

Uma biblioteca Go robusta e flexível para implementar arquiteturas multitenancy em aplicações web, com suporte a múltiplos bancos de dados, cache Redis e middlewares para frameworks populares.

## 🚀 Características

- **Suporte a Múltiplos Bancos de Dados**: PostgreSQL e MongoDB
- **Cache Distribuído**: Integração com Redis para cache de tenants
- **Middlewares HTTP**: Suporte nativo para Gin, Fiber e Chi
- **Gerenciamento de Conexões**: Pool de conexões otimizado por tenant
- **Context-Aware**: Propagação automática de informações do tenant via context
- **Observabilidade**: Integração com OpenTelemetry para tracing
- **Configuração Flexível**: Configuração via variáveis de ambiente
- **Testes Abrangentes**: Cobertura completa com testes unitários e de integração

## 📦 Instalação

```bash
go get github.com/victorximenis/multitenant
```

## 🏗️ Arquitetura

A biblioteca segue os princípios de Clean Architecture com as seguintes camadas:

- **Core**: Entidades de domínio e interfaces
- **Service**: Lógica de negócio
- **Infrastructure**: Implementações de repositórios e cache
- **Interfaces**: Middlewares HTTP e CLI
- **TenantContext**: Utilitários para gerenciamento de contexto

## ⚙️ Configuração

### Variáveis de Ambiente

Configure as seguintes variáveis de ambiente:

```bash
# Configuração do Banco de Dados
export MULTITENANT_DATABASE_TYPE=postgres  # ou mongodb
export MULTITENANT_DATABASE_DSN="postgres://user:password@localhost:5432/db?sslmode=disable"

# Configuração do Redis
export MULTITENANT_REDIS_URL="redis://localhost:6379"
export MULTITENANT_CACHE_TTL="5m"

# Configuração HTTP
export MULTITENANT_HEADER_NAME="X-Tenant-Id"

# Configuração de Pool de Conexões
export MULTITENANT_POOL_SIZE="10"
export MULTITENANT_MAX_RETRIES="3"
export MULTITENANT_RETRY_DELAY="1s"

# Configuração de Logging
export MULTITENANT_LOG_LEVEL="info"  # debug, info, warn, error
```

### Arquivo .env

Crie um arquivo `.env` na raiz do seu projeto:

```env
MULTITENANT_DATABASE_TYPE=postgres
MULTITENANT_DATABASE_DSN=postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable
MULTITENANT_REDIS_URL=redis://localhost:6379
MULTITENANT_HEADER_NAME=X-Tenant-Id
MULTITENANT_CACHE_TTL=5m
MULTITENANT_POOL_SIZE=10
MULTITENANT_MAX_RETRIES=3
MULTITENANT_RETRY_DELAY=1s
MULTITENANT_LOG_LEVEL=info
```

## 🚀 Uso Básico

### 1. Inicialização do Cliente

```go
package main

import (
    "context"
    "log"
    
    "github.com/victorximenis/multitenant"
)

func main() {
    ctx := context.Background()
    
    // Carregar configuração das variáveis de ambiente
    config, err := multitenant.LoadConfigFromEnv()
    if err != nil {
        log.Fatalf("Erro ao carregar configuração: %v", err)
    }
    
    // Criar cliente multitenant
    client, err := multitenant.NewMultitenantClient(ctx, config)
    if err != nil {
        log.Fatalf("Erro ao criar cliente multitenant: %v", err)
    }
    defer client.Close(ctx)
    
    // Usar o cliente...
}
```

### 2. Integração com Gin

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "github.com/victorximenis/multitenant"
    "github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
    ctx := context.Background()
    
    // Configurar cliente
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(ctx, config)
    defer client.Close(ctx)
    
    // Configurar Gin
    router := gin.Default()
    
    // Adicionar middleware de tenant
    router.Use(client.GinMiddleware())
    
    // Rota que usa informações do tenant
    router.GET("/api/tenant-info", func(c *gin.Context) {
        tenant, ok := tenantcontext.GetTenant(c.Request.Context())
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant não encontrado"})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "tenant_id":   tenant.ID,
            "tenant_name": tenant.Name,
            "is_active":   tenant.IsActive,
            "metadata":    tenant.Metadata,
        })
    })
    
    router.Run(":8080")
}
```

### 3. Integração com Fiber

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/victorximenis/multitenant"
    "github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
    // Configurar cliente
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(context.Background(), config)
    
    // Configurar Fiber
    app := fiber.New()
    
    // Adicionar middleware de tenant
    app.Use(client.FiberMiddleware())
    
    app.Get("/api/tenant-info", func(c *fiber.Ctx) error {
        tenant, ok := tenantcontext.GetTenant(c.Context())
        if !ok {
            return c.Status(500).JSON(fiber.Map{"error": "tenant não encontrado"})
        }
        
        return c.JSON(fiber.Map{
            "tenant_id":   tenant.ID,
            "tenant_name": tenant.Name,
            "is_active":   tenant.IsActive,
        })
    })
    
    app.Listen(":8080")
}
```

### 4. Integração com Chi

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/victorximenis/multitenant"
    "github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
    // Configurar cliente
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(context.Background(), config)
    
    // Configurar Chi
    r := chi.NewRouter()
    
    // Adicionar middleware de tenant
    r.Use(client.ChiMiddleware())
    
    r.Get("/api/tenant-info", func(w http.ResponseWriter, r *http.Request) {
        tenant, ok := tenantcontext.GetTenant(r.Context())
        if !ok {
            http.Error(w, "tenant não encontrado", http.StatusInternalServerError)
            return
        }
        
        // Responder com informações do tenant...
    })
    
    http.ListenAndServe(":8080", r)
}
```

## 🏢 Gerenciamento de Tenants

### Estrutura do Tenant

```go
type Tenant struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    IsActive    bool                   `json:"is_active"`
    Metadata    map[string]interface{} `json:"metadata"`
    Datasources []Datasource           `json:"datasources"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type Datasource struct {
    ID        string                 `json:"id"`
    TenantID  string                 `json:"tenant_id"`
    DSN       string                 `json:"dsn"`
    Role      string                 `json:"role"` // read, write, rw
    PoolSize  int                    `json:"pool_size"`
    Metadata  map[string]interface{} `json:"metadata"`
    CreatedAt time.Time              `json:"created_at"`
    UpdatedAt time.Time              `json:"updated_at"`
}
```

### Operações com Tenants

```go
// Obter serviço de tenant
tenantService := client.GetTenantService()

// Criar novo tenant
tenant := core.NewTenant("meu-tenant")
tenant.Metadata["plan"] = "premium"
err := tenantService.CreateTenant(ctx, tenant)

// Buscar tenant por nome
tenant, err := tenantService.GetTenant(ctx, "meu-tenant")

// Listar todos os tenants
tenants, err := tenantService.ListTenants(ctx)

// Atualizar tenant
tenant.IsActive = false
err = tenantService.UpdateTenant(ctx, tenant)

// Deletar tenant
err = tenantService.DeleteTenant(ctx, tenant.ID)
```

## 🔌 Gerenciamento de Conexões

### Pool de Conexões PostgreSQL

```go
// Obter gerenciador de conexões
connManager := client.GetConnectionManager()

// Obter pool PostgreSQL para leitura/escrita
pool, err := connManager.GetPostgresPool(ctx, "rw")
if err != nil {
    log.Fatal(err)
}

// Usar o pool
rows, err := pool.Query(ctx, "SELECT * FROM users WHERE tenant_id = $1", tenantID)
```

### Pool de Conexões MongoDB

```go
// Obter pool MongoDB
mongoPool, err := connManager.GetMongoPool(ctx, "read")
if err != nil {
    log.Fatal(err)
}

// Usar o pool
collection := mongoPool.Database("mydb").Collection("users")
cursor, err := collection.Find(ctx, bson.M{"tenant_id": tenantID})
```

## 🧪 Context e Utilitários

### Trabalhando com Context

```go
import "github.com/victorximenis/multitenant/tenantcontext"

// Adicionar tenant ao context
ctx = tenantcontext.WithTenant(ctx, tenant)

// Obter tenant do context
tenant, ok := tenantcontext.GetTenant(ctx)

// Obter tenant (com panic se não encontrado)
tenant := tenantcontext.MustGetTenant(ctx)

// Obter apenas o nome do tenant
tenantName := tenantcontext.GetCurrentTenantName(ctx)

// Obter apenas o ID do tenant
tenantID := tenantcontext.GetCurrentTenantID(ctx)

// Verificar se há tenant no context
hasTenant := tenantcontext.HasTenant(ctx)
```

### Utilitários para Testes

```go
import "github.com/victorximenis/multitenant/tenantcontext"

// Criar context de teste com tenant
ctx := tenantcontext.CreateTestContext("test-tenant")

// Criar tenant de teste
tenant := tenantcontext.CreateTestTenant("test-tenant", map[string]interface{}{
    "plan": "test",
})
```

## 📊 Observabilidade

### Tracing com OpenTelemetry

A biblioteca automaticamente propaga informações do tenant para spans de tracing:

```go
import "github.com/victorximenis/multitenant/tenantcontext"

// O middleware automaticamente chama:
tenantcontext.PropagateToSpan(ctx)

// Isso adiciona os seguintes atributos ao span:
// - tenant.id
// - tenant.name
// - tenant.is_active
```

### Logging

Configure o nível de log via variável de ambiente:

```bash
export MULTITENANT_LOG_LEVEL=debug  # debug, info, warn, error
```

## 🧪 Testes

### Executar Testes Unitários

```bash
go test ./...
```

### Executar Testes de Integração

Os testes de integração requerem PostgreSQL e Redis rodando:

```bash
# Iniciar dependências com Docker
docker run -d --name postgres-test -p 5432:5432 \
  -e POSTGRES_USER=dev_user \
  -e POSTGRES_PASSWORD=dev_password \
  -e POSTGRES_DB=multitenant_db \
  postgres:15

docker run -d --name redis-test -p 6379:6379 redis:7

# Executar testes de integração
go test -v ./... -tags=integration
```

### Cobertura de Testes

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Exemplos Completos

Veja o diretório [`examples/`](./examples/) para exemplos completos de uso com diferentes frameworks.

### Executar Exemplo

```bash
cd examples
export MULTITENANT_DATABASE_DSN="postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable"
export MULTITENANT_REDIS_URL="redis://localhost:6379"
go run main.go
```

### Testar Exemplo

```bash
# Health check
curl http://localhost:8080/api/health

# Com header de tenant
curl -H "X-Tenant-Id: test-tenant" http://localhost:8080/api/tenant

# Sem header (deve retornar erro)
curl http://localhost:8080/api/tenant
```

## 🔧 Configuração Avançada

### Configuração Personalizada

```go
config := &multitenant.Config{
    DatabaseType: multitenant.PostgreSQL,
    DatabaseDSN:  "postgres://...",
    RedisURL:     "redis://localhost:6379",
    HeaderName:   "X-Custom-Tenant",
    CacheTTL:     10 * time.Minute,
    PoolSize:     20,
    MaxRetries:   5,
    RetryDelay:   2 * time.Second,
    LogLevel:     "debug",
}

client, err := multitenant.NewMultitenantClient(ctx, config)
```

### Middleware Personalizado

```go
// Gin com tratamento de erro personalizado
middleware := httpMiddleware.TenantMiddleware(httpMiddleware.GinMiddlewareConfig{
    TenantService: tenantService,
    HeaderName:    "X-Tenant-Id",
    ErrorHandler: func(c *gin.Context, err error) {
        // Tratamento personalizado de erro
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Tenant inválido",
            "code":  "INVALID_TENANT",
        })
        c.Abort()
    },
})
```

## 🤝 Contribuindo

1. Fork o projeto
2. Crie uma branch para sua feature (`git checkout -b feature/nova-feature`)
3. Commit suas mudanças (`git commit -am 'Adiciona nova feature'`)
4. Push para a branch (`git push origin feature/nova-feature`)
5. Abra um Pull Request

### Diretrizes de Desenvolvimento

- Mantenha cobertura de testes acima de 80%
- Siga as convenções de código Go
- Adicione documentação para novas funcionalidades
- Execute `go fmt` e `go vet` antes de commitar

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🆘 Suporte

- 🐛 Issues: [GitHub Issues](https://github.com/victorximenis/multitenant/issues)
- 📖 Documentação: [Wiki](https://github.com/victorximenis/multitenant/wiki)

## 🗺️ Roadmap

- [ ] Suporte a MySQL
- [ ] Middleware para Echo framework
- [ ] Métricas com Prometheus
- [ ] CLI para gerenciamento de tenants
- [ ] Migração automática de schemas
- [ ] Suporte a sharding horizontal

---

**Desenvolvido com ❤️ em Go** 