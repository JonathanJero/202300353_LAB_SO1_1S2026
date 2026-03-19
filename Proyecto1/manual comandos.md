# GUIA COMANDOS PROYECTO SOPES 1

**Proyecto:** Virtualización Anidada & Microservicios (USAC SO1) 
**Arquitectura:** Mac M2 (ARM64) -> UTM -> QEMU
**Estudiante:** 202300353 Jonathan Jeronimo

---

##  CAPÍTULO 1: EL DESPERTAR (Protocolo de Inicio)



1. **Host Físico:** Abre **UTM** e inicia la **Ubuntu Desktop ("Madre")**.
2. **Host Ubuntu:** Abre la terminal en Ubuntu Madre.
3. **Encendido de vms:**
```bash
# Primero el Registro (VM3)
virsh start vm3
virsh start vm1
virsh start vm2

```


4. **Verificación de Identidad (IPs):**
 Verifica si las IPs siguen siendo las mismas.
```bash
virsh net-dhcp-leases default
# o si falla, ir en cada una haciendo ip addr
```

5. **Enlace Neural (SSH):**
Abre 3 pestañas en tu terminal y conéctate a cada una:
* `ssh vm1@192.168.122.49` (VM1)
* `ssh vm1@192.168.122.5`  (VM2)
* `ssh vm1@192.168.122.12` (VM3)



---

##  CAPÍTULO 2: EL ALMACÉN (VM3 - Registro Zot)

*Ejecutar en la terminal de VM3.*

**1. Verificar estado:**

```bash 
#en vm3
sudo docker ps

```

* Si ves `zot-registry` -> **Listo.**
* Si no hay nada -> Ejecuta:

**2. Comando de Invocación (Solo si no está corriendo):**

```bash
sudo docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name zot-registry \
  ghcr.io/project-zot/zot-linux-arm64:latest

```

**3. Verificar Inventario (Opcional):**
¿Qué imágenes tienes guardadas?

```bash
curl http://localhost:5000/v2/_catalog

```

---

##  CAPÍTULO 3: LA FÁBRICA (Ubuntu Madre - Build & Push)

*Ejecutar en Ubuntu Madre. Solo haz esto si modificaste el código.*


**Para API 1:**

```bash
# en tarminal madre
cd ~/proyectos/api1
docker build -t 192.168.122.12:5000/api1-202300353:v1 .
docker push 192.168.122.12:5000/api1-202300353:v1

```

**Para API 2:**

```bash
# en tarminal madre
cd ~/proyectos/api2
docker build -t 192.168.122.12:5000/api2-202300353:v1 .
docker push 192.168.122.12:5000/api2-202300353:v1

```

**Para API 3:**

```bash
# en tarminal madre
cd ~/proyectos/api3
docker build -t 192.168.122.12:5000/api3-202300353:v1 .
docker push 192.168.122.12:5000/api3-202300353:v1

```

---

##  CAPÍTULO 4: EL DESPLIEGUE "FÉNIX" (VM1 y VM2)

*Aquí solucionamos el problema del `snapshot exists`. Usaremos comandos combinados que limpian lo viejo antes de crear lo nuevo.*

###  EN LA VM1 (IP .49)

**1. Desplegar API 1 (Puerto 8080)**
*Este comando: Borra el snapshot viejo (si existe) -> Descarga la imagen -> Ejecuta la nueva.*

```bash
# COMANDO COMBO API1, ejecutar en vm1
sudo ctr container delete api1-viva 2>/dev/null; \
sudo ctr image pull --plain-http 192.168.122.12:5000/api1-202300353:v1 && \
sudo ctr run -d --net-host 192.168.122.12:5000/api1-202300353:v1 api1-viva

```

**2. Desplegar API 2 (Puerto 8081)**

```bash
# COMANDO COMBO API2, ejecutar en vm1
sudo ctr container delete api2-viva 2>/dev/null; \
sudo ctr image pull --plain-http 192.168.122.12:5000/api2-202300353:v1 && \
sudo ctr run -d --net-host 192.168.122.12:5000/api2-202300353:v1 api2-viva

```

###  EN LA VM2 (IP .5)

**3. Desplegar API 3 (Puerto 8080)**

```bash
# COMANDO COMBO API3, ejecutar en vm2
sudo ctr container delete api3-viva 2>/dev/null; \
sudo ctr image pull --plain-http 192.168.122.12:5000/api3-202300353:v1 && \
sudo ctr run -d --net-host 192.168.122.12:5000/api3-202300353:v1 api3-viva

```

> **¿Qué hacen las banderas?**
> * `2>/dev/null`: Silencia el error si no había nada que borrar (para que no te asuste).
> * `--net-host`: Conecta la API directo a la tarjeta de red de la VM (sin aislar puertos).
> * `--plain-http`: Permite descargar de Zot sin certificado SSL (inseguro).
> 
> 

---

##  CAPÍTULO 5: LA VALIDACIÓN (Pruebas Finales)


### 1. Pruebas Locales (Desde VM1)

Verificar que API1 y API2 están vivas.

```bash
# desde vm1
# Health API1
curl http://localhost:8080/health
# [cite_start]Esperado: {"status":"UP", "VM":"VM1", "carnet":"202300353"} [cite: 190]

# Health API2
curl http://localhost:8081/health
# Esperado: {"status":"UP", "VM":"VM1", ...}

```

### 2. Prueba de Comunicación Desde VM1


```bash
# desde vm1
curl http://localhost:8081/api2/202300353/call-api1
curl http://localhost:8080/api1/202300353/call-api2
curl http://localhost:8080/api1/202300353/call-api3
curl http://localhost:8081/api2/202300353/call-api3

```

### 3. Prueba de Comunicación desde VM2



```bash
#desde vm2
curl http://localhost:8080/api3/202300353/call-api1
curl http://localhost:8080/api3/202300353/call-api2

```

---

##  CAPÍTULO 6: EL PROTOCOLO DE CIERRE (Apagado Seguro)

**1. Apagar las Hijas (Desde Ubuntu Madre):**

```bash
virsh shutdown vm1
virsh shutdown vm2
virsh shutdown vm3

```

**2. Verificar Silencio:**
Espera unos 10 segundos y ejecuta:

```bash
virsh list --all

```

**3. Apagar la Madre:**
En la terminal de Ubuntu Madre:

```bash
sudo poweroff

```

**4. Apagar el Host**

