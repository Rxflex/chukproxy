package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v2"
)

// Структура для конфигурации базы данных
type Config struct {
	Database struct {
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DBName   string `yaml:"dbname"`
	} `yaml:"database"`
}

type Route struct {
	TargetIP   string
	TargetPort int
	Protocol   string // "tcp", "udp", or "both"
}

var (
	db           *sql.DB
	routes       = make(map[int]Route)
	listeners    = make(map[int]net.Listener)
	udpListeners = make(map[int]*net.UDPConn)
	activeLock   sync.Mutex
	config       Config
)

func loadConfig() error {
	// Открываем и читаем конфигурационный файл
	file, err := os.Open("config.yaml")
	if err != nil {
		return fmt.Errorf("Ошибка при открытии конфигурационного файла: %v", err)
	}
	defer file.Close()

	// Декодируем YAML в структуру config
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return fmt.Errorf("Ошибка при чтении конфигурационного файла: %v", err)
	}
	return nil
}

func loadRoutes() (map[int]Route, error) {
	rows, err := db.Query("SELECT listen_port, target_ip, target_port, protocol FROM routes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	newRoutes := make(map[int]Route)
	for rows.Next() {
		var port int
		var ip string
		var tport int
		var protocol string
		if err := rows.Scan(&port, &ip, &tport, &protocol); err != nil {
			log.Printf("Ошибка парсинга строки: %v", err)
			continue
		}
		newRoutes[port] = Route{TargetIP: ip, TargetPort: tport, Protocol: protocol}
	}
	return newRoutes, nil
}

func startTCPListener(port int, route Route) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Printf("Ошибка запуска TCP-порта %d: %v", port, err)
		return
	}

	activeLock.Lock()
	listeners[port] = ln
	activeLock.Unlock()

	log.Printf("✅ Слушаю TCP на порту %d → %s:%d", port, route.TargetIP, route.TargetPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("⛔️ Ошибка соединения на TCP порту %d: %v", port, err)
			break
		}
		go handleTCPConnection(conn, route)
	}
	log.Printf("⛔️ Прослушка TCP на порту %d остановлена", port)
}

func startUDPListener(port int, route Route) {
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Ошибка запуска UDP-порта %d: %v", port, err)
		return
	}

	activeLock.Lock()
	udpListeners[port] = conn
	activeLock.Unlock()

	log.Printf("✅ Слушаю UDP на порту %d → %s:%d", port, route.TargetIP, route.TargetPort)

	buf := make([]byte, 1024)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("⛔️ Ошибка получения данных с UDP порта %d: %v", port, err)
			break
		}

		go handleUDPConnection(conn, buf[:n], addr, route)
	}
	log.Printf("⛔️ Прослушка UDP на порту %d остановлена", port)
}

func stopTCPListener(port int) {
	activeLock.Lock()
	defer activeLock.Unlock()

	if ln, ok := listeners[port]; ok {
		ln.Close()
		delete(listeners, port)
		log.Printf("🛑 Остановлен TCP порт %d", port)
	}
}

func stopUDPListener(port int) {
	activeLock.Lock()
	defer activeLock.Unlock()

	if conn, ok := udpListeners[port]; ok {
		conn.Close()
		delete(udpListeners, port)
		log.Printf("🛑 Остановлен UDP порт %d", port)
	}
}

func handleTCPConnection(conn net.Conn, route Route) {
	defer conn.Close()

	targetAddr := fmt.Sprintf("%s:%d", route.TargetIP, route.TargetPort)
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		log.Printf("Ошибка подключения к %s: %v", targetAddr, err)
		return
	}
	defer targetConn.Close()

	go io.Copy(targetConn, conn)
	io.Copy(conn, targetConn)
}

func handleUDPConnection(conn *net.UDPConn, data []byte, addr *net.UDPAddr, route Route) {
	targetAddr := fmt.Sprintf("%s:%d", route.TargetIP, route.TargetPort)
	remoteAddr, err := net.ResolveUDPAddr("udp", targetAddr)
	if err != nil {
		log.Printf("Ошибка разрешения адреса для UDP: %v", err)
		return
	}

	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		log.Printf("Ошибка отправки данных на %s: %v", targetAddr, err)
	}
}

func syncRoutes() {
	newRoutes, err := loadRoutes()
	if err != nil {
		log.Printf("❌ Ошибка загрузки маршрутов: %v", err)
		return
	}

	activeLock.Lock()
	defer activeLock.Unlock()

	// Добавляем новые порты
	for port, route := range newRoutes {
		if _, exists := listeners[port]; !exists && (route.Protocol == "tcp" || route.Protocol == "both") {
			routes[port] = route
			go startTCPListener(port, route)
		}
		if _, exists := udpListeners[port]; !exists && (route.Protocol == "udp" || route.Protocol == "both") {
			routes[port] = route
			go startUDPListener(port, route)
		}
	}

	// Удаляем удалённые порты
	for port := range listeners {
		if _, exists := newRoutes[port]; !exists {
			stopTCPListener(port)
			delete(routes, port)
		}
	}
	for port := range udpListeners {
		if _, exists := newRoutes[port]; !exists {
			stopUDPListener(port)
			delete(routes, port)
		}
	}
}

func main() {
	// Чтение конфигурации из файла
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Формирование строки подключения из конфигурации
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.DBName,
	)

	var errConnecting error
	db, errConnecting = sql.Open("mysql", connStr)
	if errConnecting != nil {
		log.Fatal("❌ Ошибка подключения к БД:", errConnecting)
	}
	defer db.Close()

	// Первая синхронизация
	syncRoutes()

	// Обновление каждые 10 секунд
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		syncRoutes()
	}
}
