package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
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


type RAMInfo struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
}

type Process struct {
	PID    int    `json:"pid"`
	Nombre string `json:"nombre"`
	VSZ    uint64 `json:"vsz"`
	RSS    uint64 `json:"rss"`
	CPU    uint64 `json:"cpu"`
}

type SystemInfo struct {
	RAM      RAMInfo   `json:"ram"`
	Procesos []Process `json:"procesos"`
}

var ctx = context.Background()
var contenedoresEliminadosTotales int64 = 0

func iniciarServicios() {
	fmt.Println("Iniciando servicios automáticos...")


	exec.Command("sudo", "insmod", "/home/os1/proyecto2/continfo.ko").Run()
	fmt.Println("Módulo de Kernel cargado.")


	cronCmd := fmt.Sprintf("(crontab -l 2>/dev/null; echo \"*/2 * * * * /home/os1/proyecto2/generador.sh >> /home/os1/proyecto2/cron.log 2>&1\") | crontab -")
	exec.Command("bash", "-c", cronCmd).Run()
	fmt.Println("Cronjob de generación iniciado.")
}


func detenerServicios() {
	fmt.Println("\nDeteniendo el Daemon y limpiando servicios...")


	exec.Command("bash", "-c", "crontab -r").Run()
	fmt.Println("Cronjob eliminado con éxito.")

	exec.Command("sudo", "rmmod", "continfo").Run()
	fmt.Println("Módulo de Kernel retirado.")

	os.Exit(0)
}

func main() {
	fmt.Println("Daemon Gestor de Contenedores SO1 - 202308227")

	iniciarServicios()


	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		detenerServicios()
	}()


	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error conectando a Valkey: %v", err)
	}
	fmt.Println("Conectado a Valkey. Vigilancia activa...")


	for {
		fmt.Println("\n--- Ciclo de Análisis ---")
		analizarMemoria(rdb)
		gestionarContenedores(rdb)
		fmt.Println("Esperando siguiente ciclo...")
		time.Sleep(20 * time.Second)
	}
}

func analizarMemoria(rdb *redis.Client) {

	data, err := ioutil.ReadFile("/proc/continfo_pr2_so1_202308227")
	if err != nil {
		fmt.Println("Error: No se pudo leer el archivo /proc.")
		return
	}

	var sysInfo SystemInfo
	if err := json.Unmarshal(data, &sysInfo); err != nil {
		return
	}


	rdb.Set(ctx, "ram_total", sysInfo.RAM.Total, 0)
	rdb.Set(ctx, "ram_used", sysInfo.RAM.Used, 0)
	rdb.Set(ctx, "ram_free", sysInfo.RAM.Free, 0)


	rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "historial_ram_stream",
	  Values: map[string]interface{}{"ram_usada": sysInfo.RAM.Used},
	})


	actualizarTop5(rdb, sysInfo.Procesos)
}

func actualizarTop5(rdb *redis.Client, procesos []Process) {

	sort.Slice(procesos, func(i, j int) bool { return procesos[i].RSS > procesos[j].RSS })
	rdb.Del(ctx, "top5_ram")

	limite := 5
	if len(procesos) < 5 { limite = len(procesos) }

	for i := 0; i < limite; i++ {
		p := procesos[i]
		rdb.HSet(ctx, "top5_ram", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.RSS)
	}


	sort.Slice(procesos, func(i, j int) bool { return procesos[i].CPU > procesos[j].CPU })
	rdb.Del(ctx, "top5_cpu")
	for i := 0; i < limite; i++ {
		p := procesos[i]
		rdb.HSet(ctx, "top5_cpu", fmt.Sprintf("%d-%s", p.PID, p.Nombre), p.CPU)
	}
}

func gestionarContenedores(rdb *redis.Client) {

	out, err := exec.Command("docker", "ps", "--no-trunc", "--format", "{{.ID}}|{{.Image}}|{{.Command}}").Output()
	if err != nil { return }

	lineas := strings.Split(strings.TrimSpace(string(out)), "\n")
	var bajoConsumo, altoConsumo []string

	for _, linea := range lineas {
		if linea == "" { continue }
		partes := strings.Split(linea, "|")
		id, imagen, comando := partes[0], partes[1], partes[2]


		if strings.Contains(imagen, "grafana") || strings.Contains(imagen, "valkey") { continue }

		if strings.Contains(comando, "sleep") {
			bajoConsumo = append(bajoConsumo, id)
		} else {
			altoConsumo = append(altoConsumo, id)
		}
	}


	eliminados := 0
	for len(bajoConsumo) > 3 {
		exec.Command("docker", "rm", "-f", bajoConsumo[0]).Run()
		bajoConsumo = bajoConsumo[1:]; eliminados++
	}
	for len(altoConsumo) > 2 {
		exec.Command("docker", "rm", "-f", altoConsumo[0]).Run()
		altoConsumo = altoConsumo[1:]; eliminados++
	}

	contenedoresEliminadosTotales += int64(eliminados)
	rdb.Set(ctx, "contenedores_eliminados", contenedoresEliminadosTotales, 0)
}
