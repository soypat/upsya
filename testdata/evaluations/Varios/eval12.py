"""
<span class="badge bg-info">Desafío Final</span> Tinder para Corchetes
Desafío Final Clase 3
===
Se quiere hacer una aplicación que revise código de Python y matchee todos los corchetes. Esto quiere decir que se _pueda crear una lista valida a partir de los corchetes_. Por ejemplo: `[[],[]]` es una lista valida en python, también lo es `[],[[[[]]]]` (serían dos listas separadas).  

Para resolver este problema se recomienda emplear un _stack_. Para lograr esto se puede usar una simple lista en python que será nuestro _stack_ y empezará vacía. Usaremos solo los metodos `append(c)` y `pop()` para modificar el _stack_.
```python
corch=input()
stack = []
for c in corch:
    ...
    stack.append(c)
    ...
    stack.pop()
```

<details><summary>Ayudita</summary>

A medida que recorremos `corch` agregamos los corchetes que abren al _stack_ y cuando nos encontramos con un corchete que cierra usaremos el método `pop()`

<details><summary>Más Ayudita</summary>

Buscar en google `bracket matching` 

</details>
</details>

### Ejemplo 1
**Entrada:**  
```
[][[]]
```
Hay 3 matches. Los primeros dos corchetes se matchean entre ellos y el tercer corchete matchea con el sexto, y el cuarto con el quinto. Imprimimos `OK` para indicar que están matcheados todos los corchetes seguido de la cantidad de pares matcheados.  
**Salida:**  
```plaintext
OK 3
```
### Ejemplo 2
**Entrada:**  
```plaintext
[]][]
```
Los primeros dos corchetes están matcheados, pero el tercer corchete no tiene un par! Imprimimos `NO` seguido de la cantidad de pares matcheados antes del error.  
**Salida:**
```plaintext
NO 1
```
### Ejemplo 3
**Entrada:**
```plaintext
[[][][]
```
Este ejemplo tal vez es confuso, el error ocurre al final porque no se cerró el primer corchete   
**Salida:**
```plaintext
NO 3
```
"""
txt = input()

queue = []
matches = 0
for i,c in enumerate(txt):
   print(queue)
   if c=='[':
      queue.append(c)
   elif c==']':
      if len(queue)==0:
         print("NO", matches)
         break
      else:
         queue.pop()
         matches += 1
   if i==len(txt)-1 and len(queue)==0:
      print("OK",matches)
   elif i==len(txt)-1 and len(queue)>0:
      print("NO",matches)
"""
Stdin cases:
[][][]
---
[[]][[[]]]
---
[[][][]]
---
[]]
---
[[[[[
---
][][][]
"""
