package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

// 定义要注册的指标
var (
	cpuTemperature = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cpu_temperature_celsius",
		Help: "Current temperature of the CPU in degrees Celsius",
	})
	powerUsage = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "power_usage_watts",
		Help: "Current power usage in watts",
	})
	addr           = flag.String("listen-address", ":9010", "The address to listen on for HTTP requests.")
	scrapeInterval = flag.Int("scrape-interval", 10, "Interval between scrapes in seconds.")
)

// 初始化自定义注册表
func initCustomRegistry() *prometheus.Registry {
	reg := prometheus.NewRegistry()
	reg.MustRegister(cpuTemperature)
	reg.MustRegister(powerUsage)
	return reg
}

// 执行给定的 shell 命令并返回结果
func executeCommand(cmd string) (string, error) {
	command := exec.Command("bash", "-c", cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &out
	command.Stderr = &stderr
	err := command.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

// 使用正则提取信息
func getInfoByRegexp(input, pattern string) (string, error) {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		return match[1], nil
	}
	return "", fmt.Errorf("no info found")
}

// 记录指标的函数
func recordMetrics() {
	go func() {
		for {
			output, err := executeCommand("sensors")
			if err != nil {
				fmt.Println("Error executing command:", err)
				time.Sleep(time.Duration(*scrapeInterval) * time.Second)
				continue
			}

			// 提取 CPU 温度
			temperatureStr, err := getInfoByRegexp(output, `Package id 0:\s*\+([0-9.]+)°C`)
			if err != nil {
				fmt.Println("Error getting temperature:", err)
				time.Sleep(time.Duration(*scrapeInterval) * time.Second)
				continue
			}

			// 提取功率使用情况（假设你将来会实现这个功能）
			// powerStr, err := getInfoByRegexp(output, `PPT:\s*([0-9.]+)\s*W`)
			// if err != nil {
			//     fmt.Println("Error getting power:", err)
			//     time.Sleep(time.Duration(*scrapeInterval) * time.Second)
			//     continue
			// }

			// 解析温度值
			temperature, err := strconv.ParseFloat(temperatureStr, 64)
			if err != nil {
				fmt.Println("Error parsing temperature:", err)
				time.Sleep(time.Duration(*scrapeInterval) * time.Second)
				continue
			}
			cpuTemperature.Set(temperature)

			// 解析功率值（如果实现的话）
			// power, err := strconv.ParseFloat(powerStr, 64)
			// if err != nil {
			//     fmt.Println("Error parsing power:", err)
			//     time.Sleep(time.Duration(*scrapeInterval) * time.Second)
			//     continue
			// }
			// powerUsage.Set(power)

			// 休眠一段时间
			time.Sleep(time.Duration(*scrapeInterval) * time.Second)
		}
	}()
}

func main() {
	flag.Parse()
	reg := initCustomRegistry()
	recordMetrics()
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	fmt.Println("Starting server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
