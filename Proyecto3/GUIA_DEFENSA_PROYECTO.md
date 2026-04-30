# 🎓 Masterclass y Guía Definitiva: Defensa del Proyecto M.U.M.N.K8s

¡Bienvenido a la guía maestra de tu proyecto! Este documento está diseñado no solo para que recuerdes qué hacer en la calificación, sino para que **comprendas profundamente** cómo funciona cada engranaje del sistema. Si lees esto con atención, podrás responder cualquier pregunta técnica que te haga el auxiliar y demostrarás nivel de ingeniero Cloud/DevOps.

---

## 🛑 1. El Problema que Resuelve este Sistema

Imagina que estás en medio de un conflicto global y recibes cientos de miles de reportes por segundo de diferentes bases militares en todo el mundo indicando cuántos aviones y barcos han sido desplegados.
Si intentaras guardar todo eso directamente en una base de datos tradicional, **la base de datos colapsaría** por la cantidad masiva de conexiones y escrituras simultáneas.

**La Solución:** Una arquitectura orientada a microservicios, asíncrona y escalable. Dividimos el trabajo en pequeños programas (microservicios), usamos gRPC para que hablen entre ellos a la velocidad de la luz, colocamos un "amortiguador" (RabbitMQ) para que los datos hagan fila sin perderse, y almacenamos los resultados en una base de datos ultrarrápida en memoria (Valkey) que corre sobre Máquinas Virtuales dentro de Kubernetes (KubeVirt). Todo esto autoescalado según la demanda.

---

## 🏗️ 2. Arquitectura a Nivel de Componente (Deep Dive)

A continuación, la explicación a nivel experto de cada pieza del proyecto.

### 2.1. Locust (El Generador de Tráfico)
*   **¿Qué es?** Una herramienta de pruebas de carga escrita en Python.
*   **¿Qué hace en el proyecto?** Simula ser miles de usuarios (bases militares) enviando peticiones HTTP POST de forma concurrente (al mismo tiempo). Cada petición lleva un JSON (payload) con información de la base, cantidad de aviones y barcos.

### 2.2. Kubernetes Gateway API (El Enrutador Moderno)
*   **¿Qué es?** Es el sucesor del tradicional `Ingress`. Proporciona enrutamiento de Capa 7 (HTTP/HTTPS) hacia los servicios internos del clúster.
*   **¿Por qué usarlo?** Es más expresivo, seguro y está basado en roles. En nuestro proyecto, usamos un `Gateway` administrado por GCP (Load Balancer externo) y un `HTTPRoute` que le dice a Kubernetes: *"Todo el tráfico HTTP que entre con el prefijo `/grpc-202300353` mándalo al puerto 8080 del servicio `rust-api`"*.

### 2.3. Rust API (El Recepcionista REST)
*   **¿Qué es?** Un microservicio escrito en Rust.
*   **¿Por qué Rust?** Rust es un lenguaje compilado de altísimo rendimiento y seguridad de memoria. Como este es el primer servicio que recibe la ola masiva de tráfico de Locust, necesitamos que sea ligero y no consuma toda la RAM del clúster.
*   **¿Qué hace?** Escucha peticiones HTTP POST, extrae el cuerpo en JSON y lo reenvía internamente vía HTTP al siguiente microservicio (`Go Ingest`).

### 2.4. Go Ingest (El Traductor gRPC - Cliente)
*   **¿Qué es?** Un microservicio escrito en Golang (Go).
*   **¿Qué hace?** Recibe la petición HTTP de la API de Rust y actúa como **Cliente gRPC**.
*   **¿Por qué usamos gRPC y Protobuf?**
    *   **REST/JSON** envía la información como texto puro (strings). Ocupa mucho espacio y toma tiempo procesarlo.
    *   **gRPC/Protobuf** toma los datos, los comprime en un formato binario super compacto y los envía a través de HTTP/2 (que permite mantener la conexión abierta y enviar múltiples mensajes por el mismo "tubo"). Esto es infinitamente más eficiente para comunicación interna entre microservicios.

### 2.5. Go Writer (El Servidor gRPC)
*   **¿Qué es?** Otro microservicio en Golang.
*   **¿Qué hace?** Funciona como **Servidor gRPC**. Recibe el mensaje binario super rápido del `Go Ingest`, lo decodifica y lo inyecta (escribe) en el sistema de colas (RabbitMQ).

### 2.6. RabbitMQ (El Amortiguador / Message Broker)
*   **¿Qué es?** Un software de mensajería (Message Broker).
*   **¿Cómo funciona?** Implementa el protocolo AMQP. Tiene una "Cola" (*Queue*) llamada `wartweets`.
*   **¿Por qué es indispensable?** Es el corazón asíncrono del sistema. Si el `Go Writer` le manda 5,000 reportes en 1 segundo a RabbitMQ, pero nuestra base de datos solo soporta procesar 500 por segundo, RabbitMQ simplemente guarda los 4,500 restantes en la cola de forma segura. El consumidor los irá tomando poco a poco a su propio ritmo. Esto evita que el sistema se caiga bajo estrés.

### 2.7. Go Consumer (El Trabajador Incansable)
*   **¿Qué es?** Un microservicio en Golang.
*   **¿Qué hace?** Está suscrito a RabbitMQ. Apenas entra un mensaje nuevo a la cola `wartweets`, el consumidor lo toma, procesa los datos, y se conecta por TCP a Valkey para guardar/actualizar los contadores de barcos y aviones de cada país.

### 2.8. KubeVirt (Máquinas Virtuales Nativas en Kubernetes)
*   **El concepto clave:** Kubernetes fue diseñado para orquestar contenedores (Docker). KubeVirt es una extensión (CRD) que permite a Kubernetes **orquestar Máquinas Virtuales completas** de la misma forma en que orquesta pods.
*   **¿Cómo funciona internamente?** KubeVirt lanza un pod especial llamado `virt-launcher`. Dentro de ese pod, KubeVirt ejecuta `libvirt` y `qemu-kvm` (el mismo software que usa Linux para virtualización). ¡Así es, corre una VM real *dentro* de un contenedor!
*   **El detalle técnico que nos salvó:** Los nodos normales de GCP (`n1-standard-2`) no tienen habilitado el acceso al hardware de virtualización del procesador (el archivo `/dev/kvm`). Por lo tanto, tuvimos que parchear KubeVirt para usar **Emulación por software** (`useEmulation: true`). Es un poco más lento, pero permite que todo funcione sin requerir servidores bare-metal.

### 2.9. Valkey y Grafana (En VMs)
*   **Valkey:** Un fork de código abierto de Redis. Almacena todos sus datos en la memoria RAM, lo que hace que las lecturas y escrituras tarden menos de 1 milisegundo. Ideal para actualizar contadores en tiempo real.
*   **Grafana:** Herramienta líder para crear Dashboards y gráficos en tiempo real. Lee los datos de Valkey y los muestra visualmente.

### 2.10. Horizontal Pod Autoscaler (HPA)
*   **¿Qué hace?** Es el elástico de Kubernetes. Está monitoreando el uso de CPU de tus pods mediante el `Metrics Server`. Le indicamos que si un deployment (por ejemplo, `rust-api`) supera el 30% de uso de CPU, cree réplicas adicionales automáticamente. Cuando el ataque de Locust termine, el HPA volverá a destruir esas réplicas para ahorrar recursos.

---

## 🎬 3. Simulacro de Calificación: Comandos y Respuestas Esperadas

Sigue este guión estricto. Te muestro el comando y **lo que deberías ver y decir**.

### 🛠️ FASE 1: Verificación de Infraestructura y Zot

**1. Demuestra que Zot está funcionando fuera de Kubernetes**
*Comando:*
```bash
curl http://35.238.80.248:5000/v2/_catalog
```
*Lo que debe salir:* Un JSON listando tus repositorios: `{"repositories":["go-consumer","go-ingest","go-writer","rust-api"]}`
*Qué decir:* *"Como pueden ver, Zot está hosteado en una IP externa a mi clúster de GKE. Aquí está el catálogo de imágenes Docker que el clúster descarga para funcionar."*

**2. Gateway y Rutas**
*Comando:*
```bash
kubectl get gateway,httproute -n mumnk8s
```
*Lo que debe salir:* El gateway con estado `PROGRAMMED=True` y una dirección IP externa (ej. `35.186.255.180`). Y la ruta `mumnk8s-routes`.
*Qué decir:* *"He configurado un Gateway administrado por GCP. Ya está aprovisionado con una IP externa. El HTTPRoute está configurado para escuchar específicamente la ruta `/grpc-202300353` con mi carnet, y redirigir el tráfico a la API de Rust."*

**3. Demostrar el Autoescalado (HPA)**
*Comando:*
```bash
kubectl get hpa -n mumnk8s
```
*Lo que debe salir:* Una lista de HPA mostrando `TARGETS: X%/30%`, con `MINPODS=1` y `MAXPODS=3`.
*Qué decir:* *"Todos los deployments tienen configurado un HPA. El umbral objetivo es 30% de CPU. Si lo superamos al mandar carga con Locust, el clúster escalará dinámicamente hasta 3 pods por servicio."*

### 🛠️ FASE 2: KubeVirt (El plato fuerte)

**4. Mostrar que Valkey y Grafana son Máquinas Virtuales**
*Comandos:*
```bash
kubectl get vms,vmi -n mumnk8s
```
*Lo que debe salir:* `grafana-vm` y `valkey-vm` en STATUS `Running` y READY `True`.
*Qué decir:* *"Acá comprobamos la integración con KubeVirt. Valkey y Grafana no son pods tradicionales, son recursos del tipo VirtualMachine (VM) y VirtualMachineInstance (VMI). KubeVirt aprovisiona Persistent Volume Claims (discos) para estas VMs, permitiendo persistencia de datos tradicional sobre Kubernetes."*

### 🛠️ FASE 3: La Prueba de Carga (Tráfico Real)

**5. Enviar tráfico con Locust**
Abre una terminal nueva y ejecuta:
```bash
.\venv\Scripts\activate
locust -f locustfile.py
```
Abre tu navegador en `http://localhost:8089`.
*   **Usuarios:** 150
*   **Spawn rate:** 15
*   **Host:** `http://35.186.255.180` *(Asegúrate de pegar solo el Host base aquí sin path final si tu locustfile ya lo tiene, de lo contrario verifica)*
*Inicia la prueba.*

**6. Validar que la cadena gRPC y RabbitMQ funcionan**
Mientras Locust corre, ve a tu consola de K8s:
*Comando:*
```bash
kubectl logs deployment/rust-api -n mumnk8s --tail=5
kubectl logs deployment/go-writer -n mumnk8s --tail=5
```
*Qué decir:* *"Podemos ver en los logs cómo los servicios procesan los paquetes en tiempo real de forma síncrona mediante gRPC."*

*Comando:*
```bash
kubectl exec deployment/rabbitmq-0 -n mumnk8s -- rabbitmqctl list_queues
```
*Lo que debe salir:* `wartweets  <numero_de_mensajes>`
*Qué decir:* *"Y aquí comprobamos que los mensajes gRPC están siendo serializados y encolados exitosamente en RabbitMQ en la cola 'wartweets'."*

### 🛠️ FASE 4: Visualización
Abre Grafana en tu navegador (usa la IP de `grafana-vm`, ej: `http://136.119.0.206:3000`).
*Qué decir:* *"Finalmente, el consumidor extrae de RabbitMQ y persiste en Valkey. Grafana consulta a Valkey y renderiza los resultados. Como pueden ver, el Dashboard se actualiza en vivo con los reportes de guerra."*

---

## 🚑 4. FAQ y "Apagado de Incendios" (Troubleshooting)

Si algo sale mal o te hacen una pregunta difícil, aquí están las respuestas de experto:

**Pregunta 1: ¿Por qué no usamos Ingress estándar y elegimos Gateway API?**
**Respuesta:** "Gateway API es el nuevo estándar de Kubernetes. A diferencia de Ingress que solo soporta host y paths básicos, Gateway API es extensible, soporta enrutamiento de encabezados, gRPC nativo, y permite separar roles: un administrador gestiona el `Gateway` general y los desarrolladores gestionan el `HTTPRoute` de su aplicación."

**Pregunta 2: ¿Cómo lograste descargar imágenes de Zot si Kubernetes obliga a usar HTTPS (certificados TLS)?**
**Respuesta:** "Ese fue un reto de infraestructura. GKE bloqueaba Zot con el error `server gave HTTP response to HTTPS client`. En lugar de recompilar el `config.toml` de containerd en los nodos de GKE (lo cual es invasivo y frágil), desplegué un DaemonSet llamado `zot-proxy` en Nginx. Este proxy corre en la red del nodo (`hostNetwork`) en `localhost:5000` y proxy-pasa el tráfico a Zot. Kubernetes siempre confía en `localhost` como un entorno inseguro local, así logramos saltarnos la restricción de TLS de forma limpia e ingeniosa."

**Pregunta 3: ¿Por qué el consumidor fallaba al reiniciarse la VM de Valkey?**
**Respuesta:** "Anteriormente se estaba quemando (hardcodeando) la IP temporal que asigna Kubernetes a la VMI (`10.x.x.x`). Las VMs en KubeVirt son efímeras. Lo corregí utilizando el servicio de descubrimiento DNS interno de Kubernetes (`valkey-vm.mumnk8s.svc.cluster.local`). De esta manera, sin importar cuántas veces se reinicie la VM o cambie de IP, el DNS siempre resolverá correctamente al pod de la VM."

**Pregunta 4: ¿Qué pasa si te sale un pod en CrashLoopBackOff durante la demo?**
**Respuesta (Acción rápida):** Ejecuta `kubectl describe pod <nombre> -n mumnk8s` frente a ellos y lee la última línea de `Events`. Si el Consumer falla, explica: *"Es normal, a veces la VM de Valkey tarda un poco más en cargar el motor de Docker internamente y el Consumer falla temporalmente al conectar, pero Kubernetes lo reiniciará automáticamente (BackOff) hasta que conecte"*. (Esa es la magia de K8s, tolerancia a fallos).

---
¡Estás listo! Domina estos conceptos, confía en la arquitectura asíncrona, y recuerda que cada problema de escalabilidad que GKE y RabbitMQ resuelven de fondo, es la verdadera esencia de este laboratorio. ¡Rómpela en la calificación! 🚀
