"""
<span class="badge bg-primary">Fácil</span> Look at the time!
Convertir horas, minutos a segundos.
===
Escriba una función que dadas la hora, minutos y segundos devuelva el tiempo en segundos.
"""
def a_segundos(horas, minutos, segundos):
    return horas*60*60 + minutos*60 + segundos
"""
Placeholder:
def a_segundos(horas, minutos, segundos):
    # ...
===
Stdin cases:
2 32 2
---
0 0 0
"""
a = input().split()
print(a_segundos(int(a[0]), int(a[1]), int(a[2])))