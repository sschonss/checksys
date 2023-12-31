package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)


type SystemInfo struct {
	OS            string       `json:"os"`
	Arch          string       `json:"arch"`
	Hostname      string       `json:"hostname"`
	Time          string       `json:"time"`
	CPUUsage      float64      `json:"cpu_usage"`
	MemoryInfo    mem.VirtualMemoryStat `json:"memory_info"`
	DiskInfo      []diskInfo   `json:"disk_info"`
	Temperature   float64      `json:"temperature"`
	SystemScore   float64      `json:"system_score"`
}

type diskInfo struct {
	Mountpoint  string  `json:"mountpoint"`
	Total       uint64  `json:"total"`
	Used        uint64  `json:"used"`
	Free        uint64  `json:"free"`
	UsedPercent float64 `json:"used_percent"`
}

func main() {

	osInfo := runtime.GOOS
	archInfo := runtime.GOARCH
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println("Erro ao obter o nome do host:", err)
		return
	}

	cpuUsage, err := cpu.Percent(time.Second, false)
	if err != nil {
		fmt.Println("Erro ao obter o uso da CPU:", err)
		return
	}

	memoryInfo, err := mem.VirtualMemory()
	if err != nil {
		fmt.Println("Erro ao obter informações de memória:", err)
		return
	}

	partitions, err := disk.Partitions(true)
	if err != nil {
		fmt.Println("Erro ao obter informações de disco:", err)
		return
	}

	var diskInfoList []diskInfo
	for _, partition := range partitions {
		diskUsage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			fmt.Println("Erro ao obter informações de uso de disco:", err)
			return
		}

		diskInfoList = append(diskInfoList, diskInfo{
			Mountpoint:  partition.Mountpoint,
			Total:       diskUsage.Total,
			Used:        diskUsage.Used,
			Free:        diskUsage.Free,
			UsedPercent: diskUsage.UsedPercent,
		})
	}

	tempInfo, err := host.SensorsTemperatures()
	var temperature float64
	if err == nil && len(tempInfo) > 0 {
		temperature = tempInfo[0].Temperature
	}

	cpuScore := 100 - cpuUsage[0] 
	memScore := 100 - memoryInfo.UsedPercent	
	score := (cpuScore + memScore) / 2

	info := SystemInfo{
		OS:           osInfo,
		Arch:         archInfo,
		Hostname:     hostname,
		Time:         time.Now().Format("2006-01-02 15:04:05"),
		CPUUsage:     cpuUsage[0],
		MemoryInfo:   *memoryInfo,
		DiskInfo:     diskInfoList,
		Temperature:  temperature,
		SystemScore:  score,
	}

	jsonData, err := json.MarshalIndent(info, "", "    ")
	if err != nil {
		fmt.Println("Erro ao converter para JSON:", err)
		return
	}

	var isCritical bool
	if info.SystemScore < 50 {
		isCritical = true
	}

	var outputFile string
	if isCritical {
		outputFile = "./data/critical/" + time.Now().Format("2006-01-02_15-04-05") + ".json"
	} else {
		outputFile = "./data/" + time.Now().Format("2006-01-02_15-04-05") + ".json"
	}

	err = writeToFile(outputFile, jsonData)
	if err != nil {
		fmt.Println("Erro ao salvar as informações em um arquivo:", err)
		return
	}

	fmt.Printf("Informações do sistema salvas em %s\n", outputFile)

	if isCritical {
		err := sendEmail(outputFile)
		if err != nil {
			fmt.Println("Erro ao enviar e-mail:", err)
		} else {
			fmt.Println("E-mail enviado com sucesso.")
		}
	}
}

func sendEmail(attachmentPath string) error {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Erro ao carregar o arquivo .env:", err)
		return err
	}

	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPortStr := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	recipient := os.Getenv("RECIPIENT")
	sender := os.Getenv("SENDER")
	smtpPort, err := strconv.Atoi(smtpPortStr)
	if err != nil {
    fmt.Println("Erro ao converter a porta SMTP para inteiro:", err)
		return err
	}

	msg := "Subject: Arquivo crítico salvo\n" +
        "To: " + recipient + "\n" +
        "From: " + sender + "\n" +
        "MIME-version: 1.0;\n" +
        "Content-Type: text/html; charset=\"UTF-8\";\n\n" +
        "O arquivo crítico foi salvo e está em anexo."

    auth := smtp.PlainAuth("", smtpUsername, smtpPassword, smtpServer)

	client, err := smtp.Dial(smtpServer + ":" + strconv.Itoa(smtpPort))
	if err != nil {
		return err
	}
	defer client.Close()

	if err = client.Auth(auth); err != nil {
		return err
	}

	if err = client.Mail(sender); err != nil {
		return err
	}
	if err = client.Rcpt(recipient); err != nil {
		return err
	}

	wc, err := client.Data()
	if err != nil {
		return err
	}
	defer wc.Close()

	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		return err
	}

	attachment, err := os.Open(attachmentPath)
	if err != nil {
		return err
	}
	defer attachment.Close()

	_, err = io.Copy(wc, attachment)
	if err != nil {
		return err
	}

	return nil
}

func writeToFile(filename string, data []byte) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return err
	}

	return nil
}
