"""
<span class="badge bg-info">Desafío Final</span>  La conjetura del Dr. Lothar
Desafío Final | Clase 1
===
Escriba un programa que reciba un numero del usuario y repita el siguiente proceso usando un **while**:

*   Si el numero es par, se debe dividir por 2.
*   Si el numero es impar, se debe multiplicar por 3 y sumar 1.

Este proceso se repite hasta llegar al numero 1 y luego muestra en pantalla la cantidad de pasos que tardó en llegar.


Ejemplo:

**Entrada:**  
```plaintext
6
```

-  \\(  6, 3, 10, 5, 16, 8, 4, 2, 1 \\)  
- Se efectuaron 8 pasos.

**Salida:**  
```plaintext
8
```
**Input:** No hace falta que ustedes mismos ingresen la información de entrada del programa, el corrector realiza esto automáticamente. Alcanza con poner `n = int(input())`. **No poner ningún mensaje como parámetro del input** o sino el corrector no podrá interpretar el resultado apropiadamente.

**Output:** Deben imprimir **únicamente** el resultado en forma de un número. Al imprimir otra información en alguna parte del programa el corrector no logra interpretar el resultado apropiadamente.
"""
n = int(input())
i = 0

while n > 1:
    i += 1
    if n%2 == 0:
        n /= 2
    else:
        n = n*3 + 1

print(i)
"""
Stdin cases:
199211
"""
