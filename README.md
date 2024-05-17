# Multithreading

Segundo desafio da Pós Graduação - Go Expert da Full Cycle

- main.go

Escuta na porta 8080, na rota "/" com o parâmetro "cep" deverá consumir ao mesmo tempo duas APIs (https://brasilapi.com.br/api/cep/v1/{cep} e http://viacep.com.br/ws/{cep}/json/) e retornar a que responder primeiro, descartando a seguinte. Printa o retorno no stdout e informa qual API retornou: BrasilAPI ou ViaCEP.

Limite de tempo padrão da requisição: 1000ms.
