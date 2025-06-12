# Configuração do Serviço de Cache (Redis)

## Visão Geral

A aplicação utiliza Redis como um serviço de cache para armazenar dados frequentemente acessados, como planos de gastos de usuários, a fim de melhorar o desempenho e reduzir a carga no banco de dados principal.

## Fonte da Configuração

As configurações do Redis são gerenciadas através do arquivo de configuração principal da aplicação (e.g., `config/config.yaml`, que pode ser versionado com valores padrão ou de desenvolvimento) e podem ser sobrescritas por variáveis de ambiente. O sistema de configuração (`config/config.go`) utiliza a biblioteca Viper para carregar estas definições.

## Parâmetros de Configuração

Os seguintes parâmetros são usados para configurar a conexão com o Redis. Eles são esperados sob a chave `redis` no arquivo de configuração YAML ou através de variáveis de ambiente correspondentes.

| Chave YAML | Variável de Ambiente | Tipo   | Descrição                                                            | Exemplo         |
|------------|----------------------|--------|----------------------------------------------------------------------|-----------------|
| `address`  | `REDIS_ADDRESS`      | string | Endereço do servidor Redis (incluindo porta).                      | `localhost:6379`|
| `password` | `REDIS_PASSWORD`     | string | Senha para autenticação no servidor Redis (deixe em branco se não houver). | `""` ou `secret`|
| `db`       | `REDIS_DB`           | int    | Número do banco de dados Redis a ser usado.                          | `0`             |

## Exemplo de Configuração (YAML)

```yaml
# Exemplo em config.yaml ou um arquivo similar importado
# (A estrutura exata pode variar dependendo da configuração global do Viper,
# por exemplo, se está sob uma chave 'connections')

# Assumindo que a configuração do Redis está diretamente sob uma chave 'redis'
# no nível raiz ou dentro de um mapa específico que é carregado.
# O config/config.go espera que cfg.Redis seja populado, onde cfg é *config.Connections.
# E Connections tem: Redis RedisConfig `mapstructure:"redis"`
# Então, no YAML, esperamos uma chave 'redis':

redis:
  address: "localhost:6379"
  password: "" # ou sua senha, se houver
  db: 0        # banco de dados padrão
```

Se a estrutura do seu `config.yaml` principal tiver uma seção `connections` onde outras configurações como `firebase` residem, então `redis` também estaria lá:

```yaml
# Exemplo alternativo se aninhado em 'connections' no config.yaml
connections:
  # ... outras configurações (firebase, webserver, etc.) ...
  redis:
    address: "redis.example.com:6379"
    password: "yourredispassword"
    db: 0
```
A implementação atual em `config/config.go` (`cfg.Redis.Address`) sugere que a chave `redis` está no mesmo nível que outras configurações principais que são diretamente campos de `config.Connections`.

## Variáveis de Ambiente

As configurações do Redis podem ser fornecidas ou sobrescritas através das seguintes variáveis de ambiente. O Viper geralmente lê variáveis de ambiente no formato `PREFIX_KEY_SUBKEY`. Se nenhum prefixo específico foi configurado com `viper.SetEnvPrefix()`, ele pode procurar por `REDIS_ADDRESS`, etc., diretamente, ou pode precisar de um prefixo se a chave `redis` estiver aninhada em YAML (e.g. `CONNECTIONS_REDIS_ADDRESS`). Dado `viper.AutomaticEnv()`, ele tentará mapear.

*   `REDIS_ADDRESS`: Endereço do servidor Redis (ex: `localhost:6379`).
*   `REDIS_PASSWORD`: Senha para o servidor Redis.
*   `REDIS_DB`: Número do banco de dados Redis (ex: `0`).

É importante verificar a configuração específica do Viper na aplicação (`main.go` ou `config/config.go`) para entender como as variáveis de ambiente são prefixadas e mapeadas para as chaves YAML. A configuração atual em `config/config.go` adiciona `Redis RedisConfig `mapstructure:"redis"` diretamente à struct `Connections`, o que significa que Viper procurará por uma chave `redis` no nível superior do YAML ou variáveis como `REDIS_ADDRESS`.
