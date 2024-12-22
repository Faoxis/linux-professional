package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// Функция для сканирования одного IP
func scanIP(ip string, wg *sync.WaitGroup, results chan string) {
	defer wg.Done()

	// Пробуем подключиться к порту 80
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", ip), 500*time.Millisecond)
	if err == nil {
		results <- ip
		conn.Close()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./network_scanner <subnet> (e.g., 192.168.1.0/24)")
		os.Exit(1)
	}

	subnet := os.Args[1]

	// Парсим подсеть в формате CIDR
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		fmt.Printf("Error parsing subnet: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	results := make(chan string, 255)

	// Итерируемся по всем IP в подсети
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		wg.Add(1)
		go scanIP(ip.String(), &wg, results)
	}

	// Закрываем канал после завершения всех горутин
	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("Accessible IPs:")
	for ip := range results {
		fmt.Println(ip)
	}
}

// Утилита для инкрементации IP-адреса
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] != 0 {
			break
		}
	}
}
