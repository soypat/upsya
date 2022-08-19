"""
<span class="badge bg-info">Desafío Final</span> Perfectamente equilibrado
Desafío Final | Clase 3
===
### Paréntesis balanceados

En este desafío deben programar un [linter](https://en.wikipedia.org/wiki/Lint_(software)) que verifique la correcta utilización de los paréntesis en un texto.

La entrada del programa será un texto, que puede o no contener paréntesis `()`, corchetes `[]` y llaves `{}`, además de cualquier otra letra o símbolo. Su tarea es determinar que el texto sea válido, lo cual en este caso quiere decir que la utilización de paréntesis, corchetes y llaves es correcta, cada símbolo de apertura se corresponde con uno de cierre. Imprimir True o False si el texto es válido o no.


**Tips:**
- Investigar el comportamiento de [Pilas LIFO](https://es.wikipedia.org/wiki/Pila_(inform%C3%A1tica)) ya que son de extrema utilidad en este problema. Pueden utilizar listas de Python con los comandos `append` y `pop` para que se comporte como una pila LIFO.
- Sugerimos usar una estructura de datos para determinar las parejas de símbolos, el código será más sencillo y además será mucho más fácil agregar otras parejas de símbolos en el futuro. Algunas opciones posibles que se nos ocurrieron, aunque no las únicas, son:

```python
  openers = ['(', '{', '[']

  closers = [')', '}', ']']
```

```python
  brackets = {'(':')', '[':']', '{':'}'}
```

<details><summary>Ayudita</summary>

```python
corch=input()
stack = []
for c in corch:
    ...
    stack.append(c)
    ...
    stack.pop()
```
A medida que recorremos `corch` agregamos los corchetes que abren al _stack_ y cuando nos encontramos con un corchete que cierra usaremos el método `pop()`

<details><summary>Más Ayudita</summary>

Buscar en google `bracket matching` 

</details>
</details>

### Ejemplos de entrada/salida
El input es un texto arbitrario y la salida debe ser **únicamente** el texto `True` o `False` dependiendo de si están bien matcheados los corchetes/parentesis/llaves.

**No poner ningún mensaje como parámetro del input** o sino el corrector no podrá interpretar el resultado apropiadamente.

<details><summary>Haga click para desplegar</summary>

**Entrada:**  
  ```plaintext
Yo (Juan) quiero (necesito) café.
  ```
- Cada paréntesis se cierra:  

**Salida:**  
```python
True
```
---
**Entrada:**  
  ```plaintext
{ 1-[ 2*( 3+4 ) ] } 
  ```
- Cada símbolo se cierra en el orden correcto:

**Salida:**  
```python
True
```
---
**Entrada:**  
  ```plaintext
[ [1,2,3], [4,5,6], [7,8,9] ] 
  ```
- Cada símbolo se cierra en el orden correcto:

**Salida:**  
```python
True
```
---
**Entrada:**  
  ```plaintext
[1*(2+3)
  ```
- Falta cerrar el corchete ``]``:

**Salida:**  
```python
False
```
---
**Entrada:**  
  ```plaintext
 }[]() 
  ```
- Falta abrir la llave ``{``:

**Salida:**  
```python
False
```
---
**Entrada:**  
  ```plaintext
{ [ ( ] ) }
  ```
- Se cierran en el orden incorrecto, hay un ``]`` entre los ``( )``:

**Salida:**  
```python
False
```
---
</details>
"""

brackets = {'(':')', '[':']', '{':'}'}
def esBalanceado(corch):
   stack = []
   matched=0
   for c in corch:
      if c in brackets:
         stack.append(c)
      elif len(stack)>0  and c == brackets[stack[-1]]:
         stack.pop()
         matched += 1
      elif c in brackets.values():
         return False
   return len(stack) == 0

x = esBalanceado(input())
print(x)
"""
Stdin cases:
[][][] ok
---
][]([]) mal
---
[][][] ok
---
[[]  mal
---
()()  ok
---
[({({})})][ mal
---
[({({})})][] ok
---
[({({})})]]  mal
---
[({((){})[]})] ok
---
[({({})[]})[]]  ok
---
[({({})[]})[]][  mal

"""
