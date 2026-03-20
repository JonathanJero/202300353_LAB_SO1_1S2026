package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type RamInfo struct {
	TotalMB float64 `json:"total_mb"`
	LibreMB float64 `json:"libre_mb"`
	EnUsoMB float64 `json:"en_uso_mb"`
}

type Proceso struct {
	PID           int     `json:"pid"`
	Nombre        string  `json:"nombre"`
	VszKb         float64 `json:"vsz_kb"`
	RssKb         float64 `json:"rss_kb"`
	MemPorcentaje float64 `json:"mem_porcentaje"`
	CpuTicks      float64 `json:"cpu_ticks"`
}

type KernelData struct {
	Ram      RamInfo   `json:"ram"`
	Procesos []Proceso `json:"procesos"`
}

type Contenedor struct {
	ID      string
	Command string
	EsAlto  bool
}

var ctx = context.Background()
var totalEliminados int64 = 0

func inicializarInfraestructura() {
	fmt.Println("[*] 1. Verificando base de datos Valkey y Grafana...")
	exec.Command("sh", "-c", "cd ../db && sudo docker compose up -d").Run()
	
	fmt.Println("[*] 2. Verificando/Cargando Sonda de Kernel (Módulo C)...")
	exec.Command("sh", "-c", "lsmod | grep -q modulo || (cd ../kernel_module && sudo insmod modulo.ko)").Run()
	
	fmt.Println("[*] 3. Verificando/Programando Cronjob generador...")
	path_script, _ := os.Getwd()
	cron_cmd := fmt.Sprintf("crontab -l 2>/dev/null | grep -q 'generador.sh' || (crontab -l 2>/dev/null; echo \"*/2 * * * * %s/../cronjob/generador.sh\") | crontab -", path_script)
	exec.Command("sh", "-c", cron_cmd).Run()
	
	time.Sleep(3 * time.Second)
}

func limpiarInfraestructura() {
	fmt.Println("\n[!] Apagando daemon y limpiando sistema (Cronjob y Kernel)...")
	exec.Command("sh", "-c", "crontab -l | grep -v 'generador.sh' | crontab -").Run()
	exec.Command("sh", "-c", "lsmod | grep -q modulo && cd ../kernel_module && sudo rmmod modulo").Run()
}

func main() {
	fmt.Println("=== Iniciando Daemon Gestor ===")
	inicializarInfraestructura()

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("Error conectando a Valkey: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		limpiarInfraestructura()
		os.Exit(0)
	}()

	fmt.Printf("[%s] Infraestructura lista. Ejecutando analisis inicial inmediato...\n", time.Now().Format("15:04:05"))
	ejecutarCicloAnalisis(rdb)

	ticker := time.NewTicker(40 * time.Second)
	defer ticker.Stop()
	
	for {
		<-ticker.C
		fmt.Printf("[%s] Ejecutando analisis de contenedores...\n", time.Now().Format("15:04:05"))
		ejecutarCicloAnalisis(rdb)
	}
}

func ejecutarCicloAnalisis(rdb *redis.Client) {
	data, err := os.ReadFile("/proc/continfo_pr2_so1_202300353")
	if err != nil {
		fmt.Printf("[!] Esperando a que el modulo genere datos en /proc...\n")
		return
	}
	
	var kernelData KernelData
	json.Unmarshal(data, &kernelData)

	rdb.Set(ctx, "sys_ram_total", kernelData.Ram.TotalMB, 0)
	rdb.Set(ctx, "sys_ram_libre", kernelData.Ram.LibreMB, 0)
	rdb.Set(ctx, "sys_ram_uso", kernelData.Ram.EnUsoMB, 0)
	rdb.Set(ctx, "sys_contenedores_eliminados", totalEliminados, 0)

	sort.Slice(kernelData.Procesos, func(i, j int) bool { return kernelData.Procesos[i].RssKb > kernelData.Procesos[j].RssKb })
	rdb.Del(ctx, "top5_ram")
	for i := 0; i < 5 && i < len(kernelData.Procesos); i++ {
		p := kernelData.Procesos[i]
		rdb.HSet(ctx, "top5_ram", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.RssKb)
	}

	sort.Slice(kernelData.Procesos, func(i, j int) bool { return kernelData.Procesos[i].CpuTicks > kernelData.Procesos[j].CpuTicks })
	rdb.Del(ctx, "top5_cpu")
	for i := 0; i < 5 && i < len(kernelData.Procesos); i++ {
		p := kernelData.Procesos[i]
		rdb.HSet(ctx, "top5_cpu", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.CpuTicks)
	}

	gestionarContenedores()
}

func gestionarContenedores() {
	out, _ := exec.Command("docker", "ps", "--no-trunc", "--format", "{{.ID}}|{{.Command}}|{{.Names}}").Output()
	lineas := strings.Split(string(out), "\n")
	var altos, bajos []Contenedor

	for _, linea := range lineas {
		if linea == "" {
			continue
		}
		
		lineaMinuscula := strings.ToLower(linea)
		if strings.Contains(lineaMinuscula, "grafana") || strings.Contains(lineaMinuscula, "valkey") {
			continue
		}

		partes := strings.Split(linea, "|")
		// partes[0]=ID, partes[1]=Command, partes[2]=Name
		esAlto := strings.Contains(partes[1], "awk") || strings.Contains(partes[1], "bc")
		c := Contenedor{ID: partes[0], Command: partes[1], EsAlto: esAlto}
		
		if esAlto {
			altos = append(altos, c)
		} else {
			bajos = append(bajos, c)
		}
	}

	eliminarExceso(altos, 2, "Alto Consumo")
	eliminarExceso(bajos, 3, "Bajo Consumo")

	altosRestantes := len(altos)
	if altosRestantes > 2 { altosRestantes = 2 }
	
	bajosRestantes := len(bajos)
	if bajosRestantes > 3 { bajosRestantes = 3 }

	if altosRestantes < 2 {
		crearFaltantes(2 - altosRestantes, "Alto Consumo")
	}
	if bajosRestantes < 3 {
		crearFaltantes(3 - bajosRestantes, "Bajo Consumo")
	}
}

func eliminarExceso(lista []Contenedor, limite int, tipo string) {
	if len(lista) > limite {
		exceso := len(lista) - limite
		for i := 0; i < exceso; i++ {
			exec.Command("docker", "rm", "-f", lista[i].ID).Run()
			totalEliminados++
			fmt.Printf("   -> [X] Exceso detectado. Eliminando %s: %s\n", tipo, lista[i].ID)
		}
	}
}

func crearFaltantes(cantidad int, tipo string) {
	for i := 0; i < cantidad; i++ {
		if tipo == "Alto Consumo" {
			fmt.Printf("   -> [+] Faltan %s. Levantando uno nuevo...\n", tipo)
			exec.Command("sh", "-c", "docker run -d alpine sh -c \"while true; do echo '2^20' | bc > /dev/null; sleep 2; done\"").Run()
		} else {
			fmt.Printf("   -> [+] Faltan %s. Levantando uno nuevo...\n", tipo)
			exec.Command("sh", "-c", "docker run -d alpine sleep 240").Run()
		}
	}
}