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
Minha solução pode ser dividia em 3 pontos principais que juntos formam uma pesquisa vetorial robusta e razoavelmente rápida. É necessário informar que por mais que eu dei o meu máximo nessa rinha de backend eu acredito que minha solução não seja nem de perto competitiva usei dessa oportunidade de participar como forma de aprendizado e devo dizer que aprendi muito.

### 1. Pré processamento de dados
Como a descrição da rinha sugere os vetores não mudam durante a rinha então ficamos livre para pre-processarmos eles da forma que quisermos. o que eu decidi fazer foi transforar os vetores em JSON para 3 arquivos binários. o primeiro deles é o arquivos de vetores o vectors.bin cuja a única responsabilidade é armazenar de forma binaria os vetores para carregar todos eles em memoria mesmo.

So que carregar 3 milhões de vetores de 14 posições em memoria daria 42 milhões de posições e com cada posição do vetor sendo um float de 32 bits isso daria 1 bilhão e 344 milhões de bits que dividindo por 8 daria 168.000.000 bytes ou seja 160 Megabytes so para carregar os vetores. o que não seria possível fazer na rinha ja que existe um limite de 350 mb total. isso para duas instancias ficaria ja 320 mb sem nem botar o app pra rodar apenas carregando na memoria.

Porém os valores dos vetores vão de 0 -> 1 somente esse intervalo. com exceção dos valores sentinela como -1. Então nos podemos abstrair esses valores entre 0 -> 1 que ocupariam 32 bits para um inteiro de 8 bits. ou seja um Byte para cada posição do array. Isso so é possível pois nos temos um intervalo bem definido e por mais que fazer isso reduz a precisão, ele economiza em muito o ganho de memoria. pense que o tamanho total do arquivo é apenas 25% do original um ganho de 4X na eficiência de memoria.

Apos os vectors.bin serem criados o programa tambem cria o arquivo contendo as labels dos vetores. basicamente é uma arrayzona contendo se o vetor é fraude ou se é legitimo. Isso tambem só foi possivel pois a rinha não empoem edge cases como vetores em analise ou vetores sem analise, entao eles sempre andam juntos. 

contando de novo a memoria temos 3.000.000 vetores que tem 14 dimensões

$$
3.000.000 X 14 = 182.000.000 \\
182.000.000 de espacos

182.000.000 * 8 = 
$$

mais a array de labels que por sua vez sao em apenas 1 bit pos o vetor so pode ser 0 ou 1 fraude ou legit

E por fim no preprocessamento os vetores sao separados em buckets. Os Buckets agrupam vetores proximos baseado em alguma condicao especial. Na primeira versao do backend eu decidi usar a posicao que representava o valor da compra. por ser bem espalhada e geraria buckets mais simetricos. mas depois de algumas falhas de performance o novo meto foi usando 3 valores booleanos, como os de Compra Online, Cartao presente. ou se conhece o vendedor. A partir disso as transacoes marcadas como fraude ficaram juntas em 1 bucket permitindo assim que meu app consiga retornar mais cedo uma transacao fraudulenta.