"""
<span class="badge bg-primary">Fácil</span> I &heartsuit; JSON
Aprendamos JSON!
===
# Introducción a JSON
### 9 de cada 10 programadores aman el formato JSON

Qué sera lo que tanto agrada de este formato de datos? La flexibilidad al momento de escribirlo? La baja carga cognitiva de que todas las claves sean `strings`?

Sea lo que sea, nadie puede negar que es genial!

En [JSON](https://www.json.org/json-en.html) los objetos se rodean con llaves `{objeto}`. Por el momento se puede pensar que un objeto es igual a un diccionario de Python, pudiendo almacenar datos con claves, por ejemplo:

```json
{  "columna_1": [1,8,4,6,2,8,5],
   "columna_2": [99,56,223,301,56, 2],
   "columna_2": [0,-1,-66,-210,-333,334] }
```

El texto de arriba es código de Python _valido_ y un objeto json! Si quisieramos cargarlo a nuestro código Python bastaría con copiarlo y pegarlo a nuestro código. Pero no siempre vamos a estar nosotros para copiar y pegar el texto JSON. 

En el caso presentado a continuación se tiene un string de código JSON y precisamos [leerlo y procesarlo](https://www.w3schools.com/python/python_json.asp). Para eso usamos el modulo `json` de Python. Corran el siguiente código en Python:
```python
import json
s = '{"a":[1,2,3], "b":[4,6,8], "c":[99,98,97]}'
d= json.loads(s)
print(d["a"])
print(len(d["b"]))
print(sum(d["c"]))
```
y verán que se imprimen la \`\`columna" `a`, la longitud de `d["b"]` y la suma de `d["c"]` del diccionario.

Vamos a utilizar la libreria `json` entonces para leer el texto que nos llega por `input()`, lo cual podría ocurrir si estamos comunicandonos con un servidor.

# El desafío
Estás encargado de un servidor con millones de usuarios. 

Se te pide escribir un programa que lea el email y contraseña del usuario y se fije si existe el usuario y si coincide la contraseña.

Se tienen datos encolumnados en formato JSON que nos llegan del siguiente formato:
```json
{
	"usuarios": ["mica@mail.co", "jerry@gma.com","alber@soup.co"],
	"contra": ["abc123","caballitos","yoloswag"]
}
```
La entrada del programa son tres lineas! El programa entonces va tener tres `input()`s. La primer linea contiene el `JSON`, la segunda el `email` a verificar, y la tercera la `contraseña`. Por ende, las primeras lineas de su programa seguro sean:
```python
import json
usuarios = json.loads(input())
email = input()
password = input()
```

La salida del programa tiene que imprimir `OK` si el usuario **se encuentra en la base de datos** ***y*** **si coincide la contraseña**, imprimimos `DNE` (does not exist) si el usuario no existe y `NO` en cualquier otro caso.

### Caso ejemplo
**Entrada**:
```plaintext
{"usuarios": ["mica@mail.co","jerry@gma.com","alber@soup.co"],"contra": ["abc123","caballitos","yoloswag"]}
mica@mail.co
caballitos
```
El usuario existe y la contraseña también... pero no le corresponde la contraseña `caballitos` a mica. (la contraseña de mica sería `abc123`) por ende imprimimos:

**Salida**
```
NO
```

_Considere que no hay usuarios repetidos_.
"""
import json
jason = json.loads(input())
email = input()
password = input()
exists = False
OK = False
for i,u in enumerate(jason["usuarios"]):
    if email == u:
        exists = True
        if password == jason["contra"][i]:
            OK = True


if OK:
   print('OK')
elif exists:
   print('NO')
else:
   print('DNE')
"""
Stdin cases:
{"usuarios": ["abc","cdb","gid"],"contra": ["ip","dip","wip"]}
abc
ip
---
{"usuarios": ["abc","cdb","gid"],"contra": ["ip","dip","wip"]}
cdb
wip
---
{"usuarios": ["abc","cdb","gid"],"contra": ["ip","dip","wip"]}
SALOTRON
ip
"""
