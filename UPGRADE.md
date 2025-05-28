# Guia de Upgrade

Este documento descreve as mudanças entre versões e como migrar seu código.

## v1.0.0 (Primeira Release)

### Novidades
- ✨ Sistema de erros estruturado com códigos específicos
- ✨ Builder pattern para configuração programática
- ✨ Validação robusta de configuração
- ✨ Presets de configuração (Development, Production, Test)
- ✨ Documentação completa e exemplos
- ✨ CI/CD automatizado
- ✨ Templates para issues e PRs

### Mudanças Importantes
- **Erros**: Agora usamos `MultitenantError` com códigos específicos
- **Configuração**: Validação mais rigorosa de DSNs e URLs
- **Builder**: Nova forma programática de criar configurações

### Migração
Se você estava usando versões anteriores (desenvolvimento):

#### Tratamento de Erros
```go
// Antes
if err != nil {
    log.Printf("Erro: %v", err)
}

// Agora
if err != nil {
    if mtErr, ok := err.(*multitenant.MultitenantError); ok {
        log.Printf("Erro [%s]: %s", mtErr.Code, mtErr.Message)
        if mtErr.Cause != nil {
            log.Printf("Causa: %v", mtErr.Cause)
        }
    }
}
```

#### Configuração com Builder
```go
// Antes
config := &multitenant.Config{
    DatabaseType: "postgres",
    DatabaseDSN:  "postgres://...",
    // ...
}

// Agora (recomendado)
config := multitenant.NewConfigBuilder().
    WithDevelopmentPreset().
    WithDatabaseDSN("postgres://...").
    Build()
```

### Compatibilidade
- ✅ Totalmente compatível com Go 1.21+
- ✅ Suporte a PostgreSQL 12+
- ✅ Suporte a MongoDB 4.4+
- ✅ Suporte a Redis 6+

## Próximas Versões

### v1.1.0 (Planejado)
- Observabilidade (métricas, logging estruturado, tracing)
- Suporte a mais bancos de dados
- Performance improvements

### v2.0.0 (Futuro)
- Possíveis breaking changes na API
- Novas funcionalidades avançadas 