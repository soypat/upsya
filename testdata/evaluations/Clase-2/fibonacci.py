"""
<span class="badge bg-primary">Fácil</span> La secuencia de Filius Bonacci
No requerido | Opcional
===
Imprimir los primeros 100 números de la [sucesión de Fibonacci](https://es.wikipedia.org/wiki/Sucesi%C3%B3n_de_Fibonacci) separados por newlines. 

\\[ k _{i+1} = k _{i} + k _{i-1} \\]

donde \\( k_0 = 0, \quad k_1=1 \\)
"""
kprev=0
kcurr =1
for i in range(100):
   print(kprev)
   kprev, kcurr = kcurr, kprev+kcurr