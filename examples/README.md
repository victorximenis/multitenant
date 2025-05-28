# Multitenant Library Example

Este exemplo demonstra como usar a biblioteca multitenant com Gin.

## Configuração

Configure as seguintes variáveis de ambiente:

```bash
export MULTITENANT_DATABASE_TYPE=postgres
export MULTITENANT_DATABASE_DSN="postgres://dev_user:dev_password@localhost:5432/multitenant_db?sslmode=disable"
export MULTITENANT_REDIS_URL="redis://localhost:6379"
export MULTITENANT_HEADER_NAME="X-Tenant-Id"
export MULTITENANT_CACHE_TTL="5m"
export MULTITENANT_POOL_SIZE="10"
export MULTITENANT_LOG_LEVEL="info"
```

## Executando

```bash
cd examples
go run main.go
```

## Testando

### Health Check
```bash
curl http://localhost:8080/api/health
```

### Tenant Info (com header)
```bash
curl -H "X-Tenant-Id: test-tenant" http://localhost:8080/api/tenant
```

### Tenant Info (sem header - deve retornar erro)
```bash
curl http://localhost:8080/api/tenant
```

## Estrutura

- `main.go`: Aplicação de exemplo usando Gin
- Middleware de tenant é aplicado automaticamente
- Rotas demonstram como acessar informações do tenant no contexto
- Shutdown graceful implementado 