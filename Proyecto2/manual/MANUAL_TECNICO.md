#  Manual Técnico — Proyecto 2
**Sistemas Operativos 1 | Universidad San Carlos de Guatemala**

> **Estudiante:** Jonathan Eliud Jerónimo Salguero
> **Carnet:** 202300353
> **Curso:** Sistemas Operativos 1 — 1S2026

---

##  Tabla de Contenidos

1. [Descripción General](#1-descripción-general)
2. [Arquitectura del Sistema](#2-arquitectura-del-sistema)
3. [Requisitos Previos](#3-requisitos-previos)
4. [Estructura del Repositorio](#4-estructura-del-repositorio)
5. [Guía de Instalación y Ejecución](#5-guía-de-instalación-y-ejecución)
   - [5.1 Infraestructura (Valkey + Grafana)](#51-levantamiento-de-la-infraestructura-valkey--grafana)
   - [5.2 Módulo de Kernel](#52-carga-de-la-sonda-módulo-de-kernel-en-c)
   - [5.3 Cronjob](#53-configuración-del-generador-de-carga-cronjob)
   - [5.4 Daemon en Go](#54-ejecución-del-daemon-en-go)
6. [Verificación y Pruebas](#6-verificación-y-pruebas)
7. [Lógica de Gestión de Contenedores](#7-lógica-de-gestión-de-contenedores)
8. [Manejo de Errores y Apagado Seguro](#8-manejo-de-errores-y-apagado-seguro)
9. [Dashboard en Grafana](#9-dashboard-en-grafana)

---

## 1. Descripción General

Este proyecto implementa un sistema integral para la **monitorización proactiva, análisis automatizado y gestión inteligente de contenedores** en entornos Linux.

El sistema combina programación a **bajo nivel** (módulo de kernel en C) con **alto nivel** (Daemon en Go), resolviendo un problema real: la monitorización y estabilización autónoma de contenedores Docker sin intervención humana.

### Componentes principales

| Componente | Tecnología | Rol |
|---|---|---|
| Módulo de Kernel | C | Sensor de métricas del sistema desde el espacio del kernel |
| Daemon | Go | Cerebro del sistema: lee, decide y actúa |
| Generador de carga | Bash + Cron | Simula contenedores de prueba cada 2 minutos |
| Base de datos | Valkey (Redis) | Almacenamiento de métricas y logs |
| Dashboard | Grafana | Visualización en tiempo real |

---

## 2. Arquitectura del Sistema

```
┌─────────────────────────────────────────────────────────┐
│                  DAEMON GESTIONADOR (Go)                │
│                                                         │
│  1. Inicia contenedor Grafana                           │
│  2. Instala cronjob generador                           │
│  3. Carga módulo de kernel (script)                     │
│  4. Loop cada 20s:                                      │
│     ├── Lee /proc/continfo_pr2_so1_202300353            │
│     ├── Parsea JSON                                     │
│     ├── Decide qué contenedores eliminar                │
│     └── Escribe métricas en Valkey                      │
└──────────────────────┬──────────────────────────────────┘
                       │
       ┌───────────────┼──────────────────────┐
       ▼               ▼                      ▼
┌─────────────┐ ┌─────────────┐      ┌──────────────────┐
│   KERNEL    │ │   CRONJOB   │      │  VALKEY + GRAFANA│
│  MODULE (C) │ │  (Bash)     │      │  (Docker Compose)│
│             │ │             │      │                  │
│ task_struct │ │ Cada 2 min  │      │ Dashboard:       │
│ → /proc/    │ │ crea 5      │      │ - RAM total/libre│
│ continfo_   │ │ contenedores│      │ - Top5 CPU/RAM   │
│ pr2_so1_    │ │ aleatorios  │      │ - Eliminados     │
│ 202300353   │ │             │      │                  │
└─────────────┘ └─────────────┘      └──────────────────┘
```

### Flujo de datos

```
Kernel (task_struct)
        │
        ▼
/proc/continfo_pr2_so1_202300353  (JSON)
        │
        ▼
   Daemon Go  ──► Docker API (stop/rm contenedores)
        │
        ▼
     Valkey  ──► Grafana (Dashboard)
```

---

## 3. Requisitos Previos

### Sistema Operativo
- Ubuntu Server (arquitectura ARM64 o x86_64)
- Kernel Linux con soporte para módulos cargables

### Dependencias del sistema

```bash
# Compiladores y cabeceras del kernel
sudo apt update
sudo apt install -y build-essential linux-headers-$(uname -r)

# Entorno Go
sudo apt install -y golang-go

# Docker y Docker Compose
sudo apt install -y docker.io docker-compose-v2

# Agregar usuario al grupo docker (opcional, evita usar sudo)
sudo usermod -aG docker $USER
```

### Verificación de dependencias

```bash
# Verificar versión del kernel
uname -r

# Verificar Go
go version

# Verificar Docker
docker --version
docker compose version
```

---

## 4. Estructura del Repositorio

```
proyecto2/
├── kernel_module/
│   ├── modulo.c          # Código fuente del módulo de kernel
│   └── Makefile          # Script de compilación del módulo
├── cronjob/
│   └── generador.sh      # Script Bash generador de contenedores
├── go_daemon/
│   ├── main.go           # Código fuente del Daemon en Go
│   └── go.mod            # Dependencias del módulo Go
├── db/
│   └── docker-compose.yml # Definición de servicios Valkey + Grafana
└── README.md / MANUAL_TECNICO.md
```

---

## 5. Guía de Instalación y Ejecución

>  **Orden importante:** Siga los pasos en el orden indicado para evitar errores de conexión entre componentes.

### 5.1 Levantamiento de la Infraestructura (Valkey + Grafana)

Valkey y Grafana deben estar corriendo antes de iniciar el Daemon.

```bash
cd db
sudo docker compose up -d
```

**Verificación:**
```bash
docker ps
# Deben aparecer los contenedores de valkey y grafana corriendo
```

---

### 5.2 Carga de la Sonda (Módulo de Kernel en C)

El módulo debe compilarse con el `Makefile` provisto para que se enlace correctamente con la versión actual del kernel.

```bash
cd kernel_module

# Limpiar compilaciones previas
make clean

# Compilar el módulo
make

# Insertar el módulo en el kernel
sudo insmod modulo.ko
```

**Verificación — leer la interfaz `/proc`:**
```bash
cat /proc/continfo_pr2_so1_202300353
```

La salida esperada es un JSON con la siguiente estructura:

```json
{
  "total_ram": 8192,
  "free_ram": 4096,
  "used_ram": 4096,
  "processes": [
    {
      "pid": 1234,
      "name": "containerd",
      "cmdline": "docker-containerd",
      "vsz": 102400,
      "rss": 51200,
      "mem_percent": 0.62,
      "cpu_percent": 1.30
    }
  ]
}
```

**Para descargar el módulo:**
```bash
sudo rmmod modulo
```

---

### 5.3 Configuración del Generador de Carga (Cronjob)

>  Este paso puede omitirse si el Daemon de Go ya está configurado para instalar el cronjob automáticamente al iniciar.

```bash
cd cronjob

# Otorgar permisos de ejecución
chmod +x generador.sh

# Programar la tarea cada 2 minutos
(crontab -l 2>/dev/null; echo "*/2 * * * * $(pwd)/generador.sh") | crontab -

# Verificar que el cronjob se registró
crontab -l
```

**¿Qué hace `generador.sh`?**

Crea 5 contenedores Docker aleatorios entre las siguientes categorías:

| Categoría | Imagen | Comando | Descripción |
|---|---|---|---|
| Alto consumo RAM | `roldyoran/go-client` | `docker run -d roldyoran/go-client` | Consume RAM significativa |
| Alto consumo CPU | `alpine` | `docker run -d alpine sh -c "while true; do echo '2^20' \| bc > /dev/null; sleep 2; done"` | Alta carga de cómputo |
| Bajo consumo | `alpine` | `docker run -d alpine sleep 240` | Consumo mínimo de recursos |

---

### 5.4 Ejecución del Daemon en Go

```bash
cd go_daemon

# Descargar dependencias
go get github.com/redis/go-redis/v9

# Compilar el binario
go build -o daemon_so1 main.go

# Iniciar el servicio
./daemon_so1
```

Al iniciar, el Daemon realiza automáticamente:
1.  Levanta el contenedor de Grafana (si no está corriendo)
2.  Instala el cronjob generador de contenedores
3.  Ejecuta el script para cargar el módulo de kernel
4.  Inicia el loop principal cada 20 segundos

---

## 6. Verificación y Pruebas

### 6.1 Verificar la interfaz `/proc`

```bash
cat /proc/continfo_pr2_so1_202300353
```

Debe devolver JSON con campos: `pid`, `name`, `cmdline`, `vsz`, `rss`, `mem_percent`, `cpu_percent`.

### 6.2 Verificar la lógica de eliminación

En la terminal donde corre el Daemon, tras unos minutos deben aparecer mensajes como:

```
[*] Iteración del daemon iniciada
[+] Contenedores de alto consumo: 3  →  Eliminando 1 excedente(s)...
[!] Contenedor eliminado: abc123def456
[+] Estado final: 2 alto consumo | 3 bajo consumo
```

### 6.3 Verificar contenedores activos

```bash
docker ps
```

Se deben observar como máximo:
- 1 contenedor de Grafana
- 1 contenedor de Valkey
- 2 contenedores de alto consumo
- 3 contenedores de bajo consumo

### 6.4 Verificar escritura en Valkey

```bash
docker exec -it <nombre_contenedor_valkey> valkey-cli KEYS "*"
```

Deben aparecer claves con métricas de RAM, contenedores eliminados y procesos.

---

## 7. Lógica de Gestión de Contenedores

El Daemon aplica las siguientes **restricciones invariantes** en cada iteración:

| Restricción | Valor |
|---|---|
| Contenedores de bajo consumo activos | Siempre exactamente **3** |
| Contenedores de alto consumo activos | Siempre exactamente **2** |
| Contenedor de Grafana | **Nunca** se elimina |

### Criterios de ordenamiento para decidir qué eliminar

Cuando hay más contenedores de los permitidos, el Daemon ordena los excedentes por:

1. **Uso de RAM** (porcentaje)
2. **VSZ** — Tamaño de memoria virtual (KB)
3. **RSS** — Memoria física residente (KB)
4. **Uso de CPU** (porcentaje)

Los contenedores con mayor consumo de recursos y que excedan la cuota son detenidos (`docker stop`) y eliminados (`docker rm`).

### Ciclo de vida del loop principal

```
┌─ Cada 20 segundos ───────────────────────────────────┐
│                                                       │
│  1. Leer /proc/continfo_pr2_so1_202300353             │
│  2. Deserializar JSON                                 │
│  3. Clasificar contenedores (alto / bajo consumo)     │
│  4. Ordenar por RAM, VSZ, RSS, CPU                    │
│  5. Eliminar excedentes (stop + rm via Docker API)    │
│  6. Guardar métricas en Valkey                        │
│     ├── RAM total, libre, usada                       │
│     ├── Conteo de contenedores eliminados             │
│     └── Top 5 por CPU y RAM                           │
│                                                       │
└───────────────────────────────────────────────────────┘
```

---

## 8. Manejo de Errores y Apagado Seguro

El Daemon intercepta señales del sistema operativo (`SIGTERM`, `SIGINT`) para garantizar un cierre limpio.

**Al presionar `Ctrl + C`:**

```
^C
[!] Señal de interrupción recibida
[*] Eliminando cronjob del sistema...
[✓] Cronjob eliminado correctamente
[*] Cerrando conexión con Valkey...
[✓] Daemon detenido de forma segura
```

**Pasos para apagar el sistema completo:**

```bash
# 1. Detener el Daemon (Ctrl+C en su terminal, o:)
kill -SIGTERM $(pgrep daemon_so1)

# 2. Descargar el módulo de kernel
cd kernel_module
sudo rmmod modulo

# 3. Verificar que no quede en /proc
ls /proc/continfo_pr2_so1_202300353  # No debe existir

# 4. Bajar la infraestructura Docker
cd db
sudo docker compose down
```

---

## 9. Dashboard en Grafana

### Acceso

```
URL:      http://<IP_DE_LA_VM>:3000
Usuario:  admin
Password: admin
```

### Paneles del dashboard

El dashboard **"Panel de Contenedores — 202300353"** contiene los siguientes paneles:

| Panel | Tipo | Descripción |
|---|---|---|
| Total RAM | Card / Stat | Memoria RAM total del sistema en MB |
| Free RAM | Card / Stat | Memoria RAM libre disponible en MB |
| Total Contenedores Eliminados | Card / Stat | Acumulado histórico de eliminaciones |
| Uso de RAM en el tiempo | Time Series | Evolución del consumo de RAM con timestamps |
| Top 5 por Consumo de RAM | Pie Chart | Top 5 procesos/contenedores históricos por RAM |
| Top 5 por Consumo de CPU | Pie Chart | Top 5 procesos/contenedores históricos por CPU |
| RAM Usada | Card / Stat | Memoria RAM actualmente en uso en MB |

### Fuente de datos

- **Tipo:** Redis (compatible con Valkey)
- **Host:** `valkey:6379` (dentro de la red Docker Compose)
- **Sin autenticación** (configuración local)

---

##  Notas Adicionales

- El módulo de kernel fue desarrollado y probado sobre arquitectura **ARM64** en Ubuntu Server. Si se ejecuta en **x86_64**, recompilar con `make clean && make`.
- El porcentaje de CPU puede mostrar valores muy altos en la primera lectura debido a los cálculos diferenciales del kernel; esto es comportamiento esperado.
- El `docker-compose.yml` define una red interna para que Grafana y Valkey se comuniquen directamente sin exponer puertos innecesarios.

- Nota sobre Adaptación de Imágenes (Compatibilidad ARM64)

El enunciado del proyecto sugiere el uso de la imagen `roldyoran/go-client` para simular el perfil de "Alto Consumo" de memoria RAM. Sin embargo, el entorno de desarrollo y la máquina virtual se ejecutan sobre un procesador Apple Silicon M2 (arquitectura ARM64 / aarch64). Al intentar desplegar dicha imagen, se presentaban fallos de ejecución (salida silenciosa del contenedor) debido a que fue compilada exclusivamente para arquitecturas tradicionales `amd64` (Intel/AMD).

Para resolver esta limitación de hardware sin comprometer el rendimiento con emuladores lentos (como Rosetta), se adaptó el script generador (`generador.sh`) para utilizar la imagen oficial de `alpine`, la cual cuenta con soporte nativo para ARM64. 

Para lograr el mismo perfil de alto consumo de RAM requerido por la rúbrica, se implementó un comando interno utilizando `awk` que inunda un arreglo en memoria, logrando simular el estrés necesario de forma 100% nativa y estable:
`docker run -d alpine awk 'BEGIN {for(i=0;i<5000000;i++) a[i]="SO1_PROYECTO2_RAM_TEST_STRING_FILL"; system("sleep 240")}'`

---

