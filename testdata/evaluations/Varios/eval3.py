"""
<span class="badge bg-primary">Fácil</span>  Ordenar palabras
Manejo de strings y listas
===
Std input:
* \\(n\\) palabras separadas por espacios

Output deseado:
* Una lista de palabras ordenadas alfabeticamente enseparadas por newlines.

**Recomendación:**
Investigar el método [sort](https://www.w3schools.com/python/ref_list_sort.asp) para ordenar y [split](https://www.w3schools.com/python/ref_string_split.asp) para separar un string y convertirlo a una lista.

**Ejemplo:**  
Input:  
```plaintext
saludos al gran rey
```

Salida del programa (lo que se imprime):  
```plaintext
al
gran
rey
saludos
```
_Suponga que la entrada es toda en minusculas_

Se recomienda usar el método `.split()` sin argumentos para colimar el espacio en blanco. Ver [stack overflow](https://stackoverflow.com/questions/62013468/is-there-a-difference-between-split-vs-split) para entender la diferencia entre split con o sin argumentos.
"""
palabras = input() # aca entra el input que queremos resolver
lista = palabras.split()
lista.sort()
for pal in lista:
   print(pal)
"""
Stdin cases:
the second command mounts the device on the directory home the overall effect of excuting these two commands is that the content of the file is used as a file system rooted at the mount point

"""
