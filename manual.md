#  MANUAL TÉCNICO Y GUÍA DE INSTALACIÓN

Proyecto 1: Desarrollo, Conexión y Gestión de Contenedores en Entornos Virtualizados 

**Universidad San Carlos de Guatemala** 

**Facultad de Ingeniería - Ingeniería en Ciencias y Sistemas** 

**Curso:** Sistemas Operativos 1 

**Estudiante:** Jonathan Jeronimo - Carnet: 202300353

**Fecha:** Febrero 2026

---

## 1.  Resumen Ejecutivo y Objetivos

El presente proyecto detalla el diseño e implementación de un entorno virtualizado que integra el uso de máquinas virtuales (VMs) y contenedores, empleando tecnologías de la industria como Docker, Containerd, Go y Zot. La finalidad es simular una arquitectura distribuida realista para el alojamiento y comunicación de microservicios (APIs) de forma segura y eficiente.

**Objetivo Específico:** Identificar y configurar correctamente elementos básicos de red y virtualización, logrando un entorno 100% funcional y la comunicación cruzada entre 3 APIs desarrolladas en Go.

---

## 2.  Arquitectura del Sistema

Debido a la arquitectura del hardware anfitrión (Apple Silicon M2 - ARM64), se implementó una topología de virtualización anidada utilizando **QEMU/KVM**  sobre un host intermedio (Ubuntu Desktop).

*(El diagrama ilustra la separación de responsabilidades: ejecución vs. almacenamiento)* 

Distribución de la Infraestructura 

| Máquina Virtual | Runtime de Contenedores | Contenedor(es) Ejecutado(s) | Función Principal |
| --- | --- | --- | --- |
| **VM1** | Containerd | API1, API2 | Nodo Ejecutor Principal |
| **VM2** | Containerd | API3 | Nodo Ejecutor Secundario |
| **VM3** | Docker | ZOT | Registro Privado de Imágenes |

>  **[Espacio para captura 1: Evidencia de las 3 VMs encendidas]** > *(Inserta aquí una captura de tu terminal ejecutando `virsh list --all` donde se vean las 3 VMs en estado "running")*

---

## 3.  Herramientas y Tecnologías Utilizadas

* **Hipervisor Base:** UTM (QEMU) emulando arquitectura `aarch64`.
* **Sistemas Operativos:** Ubuntu Server 22.04 LTS (minimized).
* **Lenguaje de Programación:** Go (Golang).
* **Runtimes de Contenedores:** Docker (Gestión del registro) y Containerd (Ejecución ligera de APIs).

* **Registro OCI:** Zot Registry  para distribución privada de imágenes.



---

## 4.  Guía de Instalación y Configuración

### 4.1. Configuración del Registro Privado (VM3)

La VM3 actúa como el repositorio centralizado de nuestras imágenes Docker. Se instaló el motor de Docker  y se desplegó Zot usando el siguiente comando para exponer el puerto 5000:

```bash
docker run -d -p 5000:5000 --restart=always --name zot-registry ghcr.io/project-zot/zot-linux-arm64:latest

```

>  **[Espacio para captura 2: Evidencia del Registro Zot]** > *(Inserta aquí una captura de tu navegador o terminal ejecutando `curl http://192.168.122.12:5000/v2/` demostrando que Zot responde)*

### 4.2. Entorno de Ejecución (VM1 y VM2)

Para cumplir con los estándares de la industria, las VMs 1 y 2 no utilizan Docker, sino **Containerd** como runtime ligero.
La configuración requirió la generación del archivo `config.toml` por defecto y la configuración de Containerd para aceptar conexiones HTTP (inseguras) desde la VM3 en el momento de hacer *pull*.

>  **[Espacio para captura 3: Tareas corriendo en Containerd]** > *(Inserta aquí una captura ejecutando `sudo ctr tasks list` en la VM1, donde se vean las tareas `api1-viva` y `api2-viva` en estado RUNNING)*

---

## 5.  Desarrollo de las APIs

Se desarrollaron tres microservicios en Go. Todos exponen un endpoint base `/health` que responde con un formato JSON estructurado que indica el estado `UP`, el nombre de la VM y el carnet del estudiante.

### Flujo de Construcción (CI/CD Local)

El proceso de empaquetado para cada API fue el siguiente:

1. Compilación mediante un `Dockerfile` multietapa (usando `golang:1.22-alpine`).
2. Construcción de la imagen: `docker build -t <IP_VM3>:5000/api#-202300353:v1 .`
3. Subida al registro Zot: `docker push <IP_VM3>:5000/api#-202300353:v1`
4. Descarga y ejecución en destino mediante `ctr run --net-host`.

---

6.  Pruebas y Evidencia Funcional 

Esta sección demuestra la comunicación REST/HTTP exitosa entre los distintos servicios contenerizados.

### 6.1. Comunicación Local (Misma VM)

**Prueba:** La API2 consulta el estado de la API1 dentro de la misma máquina (VM1).

* **Comando ejecutado:** `curl http://localhost:8081/api2/202300353/call-api1`
* **Resultado esperado:** Validación de `status: UP` y respuesta de conexión exitosa.



>  **[Espacio para captura 4: Curl de API2 a API1]** > *(Inserta la captura de tu terminal mostrando la respuesta JSON exitosa con `"connection": true`)* 
> 
> 

### 6.2. Comunicación Cruzada (Distinta VM)

**Prueba:** La API3 (ubicada en la VM2) consulta el estado de la API1 (ubicada en la VM1) cruzando la red virtualizada.

* **Comando ejecutado:** `curl http://localhost:8080/api3/202300353/call-api1`

>  **[Espacio para captura 5: Curl de API3 a API1 - EL GRAN FINAL]** > *(Inserta la captura ejecutando este curl desde la VM2 y recibiendo el mensaje de éxito de la API1)*

---

## 7.  Conclusiones y Habilidades Desarrolladas

1. 
**Virtualización Eficiente:** Se demostró la viabilidad de utilizar QEMU para emular entornos de infraestructura complejos cuando la virtualización por hardware (KVM) tradicional no está disponible, optimizando recursos mediante el uso de discos dinámicos (qcow2).


2. 
**Arquitectura Desacoplada:** El uso de Zot demostró los beneficios de separar el almacenamiento de imágenes de los nodos de ejecución (Containerd), permitiendo distribuciones más rápidas.


3. 
**Habilidades Blandas:** Se practicó la resolución autónoma de problemas técnicos  (como la gestión de conflictos de *snapshots* en Containerd) y la adaptación de comandos genéricos a una red de IP dinámicas.