# Minha solucao para a rinha de backend 2026

Ola pessoal tudo bem sou Henrique Barbosa e sou um dev junior
neste readme vou demostrar minha forma de pensar/programar para a rinha de backend

## Stack e arquitetura
para a minha stack de tecnologia eu decidi usar go pois:
- consideravelmente rapido em relacao a ts que eu domino
- facilidade de deploy: por ser compilada eh so tacar o binario la e boa
- concorrencia de go: pensando no desafio da rinha eu pensei em ultilizar varias gorotines para palalelizar a busca
- era uma linguagem que eu queria ter mais experienca codando pois antes da rinha eu tinha pouquissima.

Ja a arquitetura, eu fui muito enviesado por node entao eu coloquei tudo dentro de uma pasta src, o que provou nao ser exatamente a melhor escolha para go. ja a estrutura de pastas eh algo bem simples basicamente eh tudo junto na pasta src com ressalvo dos json e de alguns serviços como o ngix


## Minha solução
para otimizar a memoria eu decidi nao usar nenhuma lib externa, o que seria bem difícil de fazer com js então o go se sobrasai de novo por ja ter uma lib http bem completa inclusa.

para a vetorizacao do payload eu apenas segui o que os readmes da rinha recomendaram, acredito que nao tenha muito como otimizar essa parte em especifico.

agora chegando na parte principal, a busca vetorial, de comeco eu pensei em fazer algo simples e ir melhorando conforme eu pudesse.

agora eu estou fazendo uma busca vetorial simples usando o quadrado das distancias, sem nenhuma otimizacao muito grande, eu aloquei 3 milhoes de arrays com dados simples float32 e fiz o parsing de forma gradual.

porem devido a limitacao de memoria, essa estrategia se tornou incapaz de se enquadrar nos limites, estou pensando em uma nova solucao, basicamente pre-processar os vetores para que eles sejam bytes. ou talvez criar algum algoritimo para transformar vetores de 14 dimensoes em vetores de dimensoes menores podendo ocupar menos dados em memoria.

apliquei algoritimos de quantizacao eficientes para floats em integers garantindo uma reducao de 450mb para apenas 40 mb aproximadamente 10x de eficiencia