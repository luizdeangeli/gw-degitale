# DEGITALE-GW — centralizador de gateways

Consulta de CNPJ e CEP em `gw.degitale.com.br`. Uma URL e um token por sistema.
A cadeia de provedores fica no `.env`: se o atual responde 429, 5xx, timeout ou
credencial inválida, a chamada segue para o próximo. 404 encerra, porque o
documento não existe.

| Item | Valor |
|------|-------|
| Host | `gw.degitale.com.br` |
| Container | `degitale-gw` |
| Project (compose) | `degitale-gw` |
| Porta interna | 8080 (não publicada no host — só o Traefik alcança) |
| Imagem | build local (`Dockerfile`, binário Go) |

## Endpoints

Autenticação: `Authorization: Bearer <token>`. `/health` é público.

| Método | Caminho | Resposta |
|--------|---------|----------|
| `GET` | `/health` | liveness |
| `GET` | `/v1/cnpj/{cnpj}` | 14 dígitos; pontos e traços são ignorados |
| `GET` | `/v1/cep/{cep}` | 8 dígitos |

O CNPJ não pode levar barra no caminho (`00.000.000/0001-00` quebra a rota).
Envie só os dígitos.

Sucesso:

```json
{
  "ok": true,
  "provider": "cnpjws",
  "cached": false,
  "data": {}
}
```

`data` de CNPJ segue o objeto do SMG (`nmEmpresarial`, `nmFantasia`,
`cdTipoEstabelecimento`, `dtAbertura`, `flOptanteSimplesNacional`,
`dsSituacaoCadastral`, `atividades`, `endereco`, `telefones`, `emails`,
`inscricoesEstaduais`). `data` de CEP segue o ViaCEP (`cep`, `logradouro`,
`complemento`, `bairro`, `localidade`, `uf`, `ibge`, `ddd`).

Erros: `401` token ausente ou inválido, `400` documento malformado,
`404` não encontrado, `503` cadeia esgotada (`tentativas` traz provedor e status).

Consulta bem-sucedida fica em memória: CNPJ 24h, CEP 7 dias. A segunda chamada
do mesmo documento não gasta cota.

## Cadeia inicial

```
CNPJ: cnpjws → cnpja
CEP:  viacep
```

`cnpjws` é a API pública (cerca de 3 consultas por minuto). `cnpja` só entra se
`CNPJA_API_KEY` estiver preenchida. Sem a chave, o processo sobe só com o que
deu para configurar e registra isso no log.

## Primeira instalação

```bash
# 1) DNS: registro A  gw → 187.77.243.144
#    Cloudflare proxied (nuvem laranja), SSL/TLS em modo Full

# 2) Na VPS
cd /opt/docker/sites/degitale.com.br/gw
cp -n .env.example .env
# edite GW_TOKENS e CNPJA_API_KEY

# 3) Subir
docker compose up -d --build
docker compose logs -f gw
```

O `bin/start.sh` pula a stack se o `.env` não existir.

Para rotacionar um token: edite `GW_TOKENS` e rode
`docker compose up -d --force-recreate`.

## Testes

```bash
curl -s https://gw.degitale.com.br/health

curl -s https://gw.degitale.com.br/v1/cnpj/47960950000121 \
  -H "Authorization: Bearer $GW_TOKEN"

curl -s https://gw.degitale.com.br/v1/cep/01001000 \
  -H "Authorization: Bearer $GW_TOKEN"
```

## Uso nos sistemas

```php
$base  = 'https://gw.degitale.com.br';
$token = getenv('GW_TOKEN'); // o segredo deste sistema em GW_TOKENS

$ch = curl_init($base . '/v1/cnpj/' . $cnpj);
curl_setopt_array($ch, [
    CURLOPT_HTTPHEADER => ['Authorization: Bearer ' . $token, 'Accept: application/json'],
    CURLOPT_RETURNTRANSFER => true,
]);
```

Cada sistema tem o seu par `nome:segredo`. O log do container mostra o nome
(`client=smg`), nunca o segredo.

## Novo provedor

1. Implemente `provider.Lookup` (`Name` + `Lookup`) devolvendo o JSON normalizado
   ou um `*provider.Error` com o tipo da falha (`KindNotFound`, `KindRateLimit`,
   `KindAuth`, `KindTemporary`).
2. Registre o nome em `buildCnpj` ou `buildCep` em `cmd/gw/main.go`.
3. Inclua o nome em `CNPJ_CHAIN` ou `CEP_CHAIN`.

## Operação

```bash
docker compose ps
docker compose logs -f --tail=200 gw
docker compose up -d --build
```

O endpoint é público. Quem tiver um token consome a cota de todos os provedores.
Trate cada segredo como senha de produção.
