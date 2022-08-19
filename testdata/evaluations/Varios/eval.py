"""
Las Naranjas de Miguel
Que no se quede sin naranjas el Migue
===

Miguel vive en un pueblo frutero con su hermana en el valle de Oz. 
Todos los días cosecha bananas y naranjas de su campo. Como son abundantes,
suele darle 2 bananas y 1 naranja a su hermana. No obstante Miguel siempre 
quiere quedarse con por lo menos una naranja, por lo cual le da una naranja
a su hermana solo cuando se cosechan al menos 2 naranjas.

Miguel ahora quiere modernizarse, compró una cinta transportadora que 
detecta cada fruta que la atraviesa y te pide ayuda para escribir un 
programa que reciba el código generado por la máquina y **devuelva la 
cantidad de bananas y naranjas que le quedarán a Miguel** luego de quitar 
las frutas que le dará a su hermana.

Ejemplo 1

Miguel recibe 5 bananas y 8 naranjas.

Entrada:
```plaintext
BBBBBNNNNNNNN
```

Miguel tiene suficientes naranjas para distribuirle a la hermana entonces se
queda con 3 bananas y 7 naranjas. Se imprime primero la cantidad de bananas 
y segundo la cantidad de naranjas que tiene Miguel después de convidar.  

Salida:
```plaintext
3 7
```

Ejemplo 2

Entrada:
```plaintext
BNN
```

Miguel solo le puede dar una banana a su hermana, pero tiene suficientes naranjas para los dos.

Salida:
```plaintext
0 1
```
"""

P = input()

def leerPedido(pedido):
    naranjas=0
    bananas=0
    for i in range(len(pedido)): 
        letra = pedido[i]
        if letra=="N": 
            naranjas+=1 
        elif letra=="B":
            bananas+=1
    return naranjas,bananas

def cuantasDar(naranjas,bananas):
    naranjasParaHermana = 0
    bananasParaHermana = 0
    if naranjas<2:
        naranjasParaHermana = 0
    else:
        naranjasParaHermana = 1 

    if bananas > 1:
        bananasParaHermana = 2
    else:
        bananasParaHermana = bananas

    return naranjasParaHermana,bananasParaHermana

n,b = leerPedido(P) #Leo pedido y guardo a memoria las frutas que tengo
nh,bh = cuantasDar(n,b) #Aqui calculo naranjas para darle a la hermana

print(b-bh,n-nh)

"""
Stdin cases:
BBBBBBBBN
---
BN
---
N
---
B
---

---
BBBB
---
NNNN
"""