## Rate Limit Challenge

O Rate Limit Challenge é uma aplicação web projetada para demonstrar a implementação de um sistema de limitação de taxa de requisições usando Go, Redis e Docker. Ele regula o número de requisições que um usuário pode fazer dentro de um determinado período, ajudando a proteger a aplicação contra sobrecarga ou ataques de negação de serviço (DDoS).

## Estrutura do Projeto

A aplicação é dividida em alguns pacotes, organizados da seguinte forma:

- **main.go**: O ponto de entrada da aplicação. Configura e inicia o servidor web, juntamente com o sistema de limitação de taxa.

- **config**: Carrega as configurações da aplicação a partir de variáveis de ambiente ou arquivos .env, incluindo configurações para a limitação de taxa e detalhes do servidor Redis.

- **infra/**: Contém infraestruturas auxiliares, como o servidor web (web) e a conexão com o banco de dados Redis (database).

- **database/**: Define as operações do banco de dados Redis usadas pela aplicação.

- **web**: Inclui o servidor HTTP e middlewares, como o middleware de limitação de taxa.

- **ratelimiter**: Implementa a lógica de limitação de taxa, incluindo a verificação e registro de tokens personalizados e a limitação baseada em IP.

- **handler**: Define os manipuladores de rotas HTTP para a aplicação.

## Executando a Aplicação Localmente com Docker Compose

Para rodar a aplicação localmente utilizando o Docker Compose, siga os passos abaixo:

Certifique-se de ter o Docker e o Docker Compose instalados em sua máquina. Se você ainda não os tem instalados, você pode baixá-los a partir dos seguintes links:

- Docker: [https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/)
- Docker Compose: [https://docs.docker.com/compose/install/](https://docs.docker.com/compose/install/)

Depois de instalar o Docker e o Docker Compose, siga os passos abaixo para executar a aplicação:

1. Clone o repositório da aplicação para o seu ambiente local.

2. Navegue até a pasta raiz do projeto.

3. Execute o comando abaixo para construir e iniciar os contêineres da aplicação, Redis e do Fortio.

```bash
docker-compose up --build
```
4. A aplicação estará acessível localemnte em http://localhost:8080.

## Tokens Personalizaveis

Assim que a aplicação é iniciada, no pacote config são registados 5 tokens personalizaveis, sendo eles:

**TOKEN_1**
**TOKEN_2**
**TOKEN_3**
**TOKEN_4**
**TOKEN_5**

No arquivo **.env** na raiz do projeto, é possivel pernalizar as configurações dos tokens registrados acima, controlando a quantidade de requisições que cada um poderá fazer, por segundo.

Além da quantidade de requisições por segundo de cada token, também é possivel controlar a quantidade de requisições por IP, quaque é o parametro usado quando o client não possui um token registrado no header.

As configurações nas variávereis **LOCK_DURATION_SECONDS** e **BLOCK_DURATION_SECONDS** refletem para todos os tokens e IP's. A **LOCK_DURATION_SECONDS** significa o range de tempo que usaremos para controlar a quantidade de requisições e o **BLOCK_DURATION_SECONDS** é o tempo determinado que o IP ou Token ficará impossibilitado de realizar chamadas na API.

```bash
TOKEN_1_MAX_REQUESTS_PER_SECOND=6
TOKEN_2_MAX_REQUESTS_PER_SECOND=12
TOKEN_3_MAX_REQUESTS_PER_SECOND=18
TOKEN_4_MAX_REQUESTS_PER_SECOND=24
TOKEN_5_MAX_REQUESTS_PER_SECOND=500

IP_MAX_REQUESTS_PER_SECOND=3

LOCK_DURATION_SECONDS=1
BLOCK_DURATION_SECONDS=60
```

## Exemplos de Uso
Para testar o limitador de taxa, você pode fazer solicitações HTTP para o servidor. Por exemplo, você pode usar o comando curl para fazer uma solicitação GET: curl `http://localhost:8080 -H "API_KEY: TOKEN_1"`.

Se você fizer mais solicitações do que o limite permitido em um determinado período de tempo, o servidor responderá com um código de status HTTP 429 (Too Many Requests).

## Executando Testes de Carga com Fortio
Primeiramente, certifique-se de que o serviço Fortio esteja rodando, conforme configurado no `docker-compose.yml`. A aplicação e o serviço Fortio devem estar operacionais.

**Exemplo 1: Teste de Carga**
Para executar um teste simples de carga, onde você envia requisições para a URL base da sua aplicação, use o seguinte comando:

```bash
docker exec -it <container_id> fortio load -c 2 -qps 12 -t 5s -H "API_KEY: TOKEN_2" http://app:8080
```

Neste comando:

1. <container_id> substitua esse valor pelo id do container que subiu localmente com o comando docker-compose up build. Voce encontrará essa informação rondando o comando docker container ps.
2. -c 2 especifica que 2 conexões concorrentes serão usadas.
3. -qps 12 limita as requisições por segundo a 12. Se você quiser testar sem limites, pode usar -qps 0.
4. -t 5s define a duração do teste para 5 segundos.
5. -H informa que você passará algo no Heager.
6. "API_KEY: TOKEN_2" é a chave/valor passado no Header.
7. http://app:8080/ é a URL onde o Fortio enviará as requisições. Substitua app pelo nome do serviço da aplicação, se for diferente em seu `docker-compose.yml`.

**Interpretando Resultados**
Após a execução do teste, o Fortio fornecerá um relatório detalhado, incluindo o número de requisições realizadas, a taxa de sucesso, tempos de resposta (latência), e outras métricas relevantes. Esses resultados ajudarão a entender o comportamento da sua aplicação sob carga e como a limitação de taxa afeta a experiência do usuário.

