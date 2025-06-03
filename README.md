# Multitenant Go Library

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/victorximenis/multitenant)](https://goreportcard.com/report/github.com/victorximenis/multitenant)

Uma biblioteca Go robusta e flexível para implementar arquiteturas multitenancy com suporte a múltiplos bancos de dados, cache Redis e middlewares HTTP.

## 🚀 Características

- **Multi-database**: Suporte para PostgreSQL e MongoDB
- **Cache Redis**: Cache automático com TTL configurável
- **HTTP Middlewares**: Integração com Gin, Fiber e Chi
- **CLI/Worker Support**: Resolução de tenant para aplicações não-HTTP
- **Thread-Safe**: Implementação segura para concorrência
- **Clean Architecture**: Separação clara de responsabilidades
- **Extensível**: Interfaces bem definidas para customização

## 📦 Instalação

```bash
go get github.com/victorximenis/multitenant
```

## 🏃 Quick Start

### 1. Configuração Básica

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
        log.Fatal(err)
    }
    
    // Criar cliente multitenant
    client, err := multitenant.NewMultitenantClient(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close(ctx)
    
    // Usar o cliente...
}
```

### 2. Configuração de Ambiente

```bash
export MULTITENANT_DATABASE_TYPE=postgres
export MULTITENANT_DATABASE_DSN="postgres://user:pass@localhost:5432/db?sslmode=disable"
export MULTITENANT_REDIS_URL="redis://localhost:6379"
export MULTITENANT_HEADER_NAME="X-Tenant-Id"
export MULTITENANT_CACHE_TTL="5m"
export MULTITENANT_POOL_SIZE="10"
export MULTITENANT_LOG_LEVEL="info"
export MULTITENANT_IGNORED_ENDPOINTS="/health,/metrics"  # Lista de endpoints a serem ignorados pelo middleware
```

## 🌐 Uso com HTTP Frameworks

### Gin

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
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(ctx, config)
    defer client.Close(ctx)
    
    router := gin.Default()
    
    // Adicionar middleware de tenant
    router.Use(client.GinMiddleware())
    
    router.GET("/api/tenant", func(c *gin.Context) {
        tenant, ok := tenantcontext.GetTenant(c.Request.Context())
        if !ok {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "tenant not found"})
            return
        }
        
        c.JSON(http.StatusOK, gin.H{"tenant": tenant})
    })
    
    router.Run(":8080")
}
```

### Fiber

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/victorximenis/multitenant"
    "github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
    ctx := context.Background()
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(ctx, config)
    defer client.Close(ctx)
    
    app := fiber.New()
    
    // Adicionar middleware de tenant
    app.Use(client.FiberMiddleware())
    
    app.Get("/api/tenant", func(c *fiber.Ctx) error {
        tenant, ok := tenantcontext.GetTenant(c.Context())
        if !ok {
            return c.Status(500).JSON(fiber.Map{"error": "tenant not found"})
        }
        
        return c.JSON(fiber.Map{"tenant": tenant})
    })
    
    app.Listen(":8080")
}
```

### Chi

```go
package main

import (
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/victorximenis/multitenant"
    "github.com/victorximenis/multitenant/tenantcontext"
)

func main() {
    ctx := context.Background()
    config, _ := multitenant.LoadConfigFromEnv()
    client, _ := multitenant.NewMultitenantClient(ctx, config)
    defer client.Close(ctx)
    
    r := chi.NewRouter()
    
    // Adicionar middleware de tenant
    r.Use(client.ChiMiddleware())
    
    r.Get("/api/tenant", func(w http.ResponseWriter, r *http.Request) {
        tenant, ok := tenantcontext.GetTenant(r.Context())
        if !ok {
            http.Error(w, "tenant not found", http.StatusInternalServerError)
            return
        }
        
        // Retornar tenant como JSON...
    })
    
    http.ListenAndServe(":8080", r)
}
```

## 🔧 Configuração Avançada

### Configuração Programática

```go
config := &multitenant.Config{
    DatabaseType:     multitenant.PostgreSQL,
    DatabaseDSN:      "postgres://user:pass@localhost:5432/db",
    RedisURL:         "redis://localhost:6379",
    CacheTTL:         5 * time.Minute,
    HeaderName:       "X-Tenant-Id",
    PoolSize:         10,
    MaxRetries:       3,
    RetryDelay:        1 * time.Second,
    LogLevel:         "info",
    IgnoredEndpoints: []string{"/health", "/metrics"},
}

client, err := multitenant.NewMultitenantClient(ctx, config)
```

### Configuração com MongoDB

```go
config := &multitenant.Config{
    DatabaseType: multitenant.MongoDB,
    DatabaseDSN:  "mongodb://localhost:27017/multitenant",
    RedisURL:     "redis://localhost:6379",
    // ... outras configurações
}
```

## 🖥️ Uso em CLI/Workers

### Resolver Tenant por Variável de Ambiente

```go
resolver := client.GetTenantResolver()

// Resolver tenant da variável TENANT_NAME
ctx, err := resolver.ResolveTenantFromEnv(context.Background())
if err != nil {
    log.Fatal(err)
}

// Usar contexto com tenant...
```

### Processar Todos os Tenants

```go
err := resolver.ForEachTenant(ctx, func(tenantCtx context.Context) error {
    tenant, _ := tenantcontext.GetTenant(tenantCtx)
    log.Printf("Processando tenant: %s", tenant.Name)
    
    // Sua lógica de processamento aqui...
    return nil
})
```

### Worker com Polling

```go
import "github.com/victorximenis/multitenant/interfaces/cli"

worker := cli.NewWorker(cli.WorkerConfig{
    TenantService: client.GetTenantService(),
    ProcessAll:    true, // ou false para tenant específico
    TenantName:    "specific-tenant", // se ProcessAll = false
    EnvVarName:    "TENANT_NAME",
    PollInterval:  30 * time.Second,
})

err := worker.Start(ctx, func(tenantCtx context.Context) error {
    tenant, _ := tenantcontext.GetTenant(tenantCtx)
    log.Printf("Processando tenant: %s", tenant.Name)
    return nil
})
```

## 🗄️ Gestão de Tenants

### Criar Tenant

```go
tenant := core.NewTenant("meu-tenant")
tenant.Metadata["plan"] = "premium"

err := client.GetTenantService().CreateTenant(ctx, tenant)
```

### Buscar Tenant

```go
tenant, err := client.GetTenantService().GetTenant(ctx, "meu-tenant")
if err != nil {
    log.Fatal(err)
}
```

### Listar Tenants

```go
tenants, err := client.GetTenantService().ListTenants(ctx)
```

### Adicionar Datasource ao Tenant

```go
datasource := core.NewDatasource(
    tenant.ID,
    "postgres://tenant1:pass@localhost:5432/tenant1_db",
    "rw", // read-write
    5,    // pool size
)

tenant.Datasources = append(tenant.Datasources, *datasource)
err := client.GetTenantService().UpdateTenant(ctx, tenant)
```

## 🔌 Conexões de Banco por Tenant

### PostgreSQL

```go
// Obter pool para tenant atual no contexto
pool, err := client.GetConnectionManager().GetPostgresPool(ctx, "read")

// Obter pool para tenant específico
pool, err := client.GetConnectionManager().GetPostgresPoolForTenant(ctx, "tenant-name", "write")

// Usar pool
rows, err := pool.Query(ctx, "SELECT * FROM users")
```

### MongoDB

```go
// Obter client para tenant atual
mongoClient, err := client.GetConnectionManager().GetMongoClient(ctx, "read")

// Obter client para tenant específico
mongoClient, err := client.GetConnectionManager().GetMongoClientForTenant(ctx, "tenant-name", "write")

// Usar client
collection := mongoClient.Database("mydb").Collection("users")
```

## 🧪 Testes

### Contexto de Teste

```go
import "github.com/victorximenis/multitenant/tenantcontext"

func TestMyFunction(t *testing.T) {
    // Criar tenant de teste
    tenant := tenantcontext.NewTestTenant("test-tenant")
    
    // Criar contexto com tenant
    ctx := tenantcontext.NewTestContextWithTenant(tenant)
    
    // Usar em testes...
    result := MyFunction(ctx)
    
    // Verificar se tenant está no contexto
    assert.True(t, tenantcontext.AssertTenantInContext(ctx, "test-tenant"))
}
```

## 📋 Variáveis de Ambiente

| Variável | Descrição | Padrão | Obrigatória |
|----------|-----------|---------|-------------|
| `MULTITENANT_DATABASE_TYPE` | Tipo do banco (`postgres` ou `mongodb`) | `postgres` | Não |
| `MULTITENANT_DATABASE_DSN` | String de conexão do banco | - | Sim |
| `MULTITENANT_REDIS_URL` | URL do Redis | - | Sim |
| `MULTITENANT_CACHE_TTL` | TTL do cache (ex: `5m`, `1h`) | `5m` | Não |
| `MULTITENANT_HEADER_NAME` | Nome do header HTTP | `X-Tenant-Id` | Não |
| `MULTITENANT_POOL_SIZE` | Tamanho do pool de conexões | `10` | Não |
| `MULTITENANT_MAX_RETRIES` | Máximo de tentativas | `3` | Não |
| `MULTITENANT_RETRY_DELAY` | Delay entre tentativas | `1s` | Não |
| `MULTITENANT_LOG_LEVEL` | Nível de log (`debug`, `info`, `warn`, `error`) | `info` | Não |
| `MULTITENANT_IGNORED_ENDPOINTS` | Lista de endpoints a serem ignorados pelo middleware | - | Não |

## 🚨 Troubleshooting

### Erro: "tenant not found"

**Causa**: Header `X-Tenant-Id` não enviado ou tenant não existe no banco.

**Solução**:
1. Verificar se o header está sendo enviado
2. Verificar se o tenant existe: `client.GetTenantService().GetTenant(ctx, "tenant-name")`
3. Criar o tenant se necessário

### Erro: "database DSN is required"

**Causa**: Variável `MULTITENANT_DATABASE_DSN` não configurada.

**Solução**: Configurar a variável com a string de conexão correta.

### Erro: "Redis URL is required"

**Causa**: Variável `MULTITENANT_REDIS_URL` não configurada.

**Solução**: Configurar a variável com a URL do Redis.

### Performance Issues

**Sintomas**: Lentidão nas requisições.

**Soluções**:
1. Aumentar `MULTITENANT_POOL_SIZE`
2. Ajustar `MULTITENANT_CACHE_TTL` para maior valor
3. Verificar latência do Redis
4. Monitorar conexões do banco

### Memory Leaks

**Sintomas**: Uso crescente de memória.

**Soluções**:
1. Chamar `client.Close(ctx)` ao finalizar
2. Verificar se conexões estão sendo fechadas
3. Monitorar pools de conexão

## 🤝 Contribuindo

Contribuições são bem-vindas! Por favor, leia [CONTRIBUTING.md](CONTRIBUTING.md) para detalhes sobre nosso código de conduta e processo de submissão de pull requests.

## 📄 Licença

Este projeto está licenciado sob a Licença MIT - veja o arquivo [LICENSE](LICENSE) para detalhes.

## 🔗 Links Úteis

- [Documentação da API](https://pkg.go.dev/github.com/victorximenis/multitenant)
- [Exemplos](examples/)
- [Issues](https://github.com/victorximenis/multitenant/issues)
- [Releases](https://github.com/victorximenis/multitenant/releases)

## 📊 Status do Projeto

- ✅ Core functionality
- ✅ PostgreSQL support
- ✅ MongoDB support
- ✅ Redis cache
- ✅ HTTP middlewares (Gin, Fiber, Chi)
- ✅ CLI/Worker support
- ✅ Comprehensive tests
- 🔄 Observability features (em desenvolvimento)
- 🔄 Metrics and monitoring (planejado) 