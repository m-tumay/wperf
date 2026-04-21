package main
//wperf
import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultPort = "5202"
	bufferSize  = 256 * 1024 // 256 KB
)

// ANSI Renk Kodları (Sütunlar ve çıktılar için)
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Bold   = "\033[1m"
)

var stopTest int32 = 0
var stdinLines = make(chan string)

func init() {
	prepareConsole()

	// Klavyeden girilen tüm Satır/Enter işlemlerini yakalayan global sistem.
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			stdinLines <- line
		}
	}()
}

func printHeader() {
	fmt.Printf("\n%s%s%-15s    %-15s    %-15s%s\n", Bold, Cyan, "Zaman Aralığı", "Transfer", "Bant Genişliği", Reset)
	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)
}

func printSummary(totalBytes uint64, startTime time.Time) {
	duration := time.Since(startTime).Seconds()
	if duration <= 0.001 {
		duration = 0.001 // sıfıra bölme hatasını engelle
	}
	mbytes := float64(totalBytes) / (1024 * 1024)
	mbits := (float64(totalBytes) * 8) / (1000 * 1000)
	avgMbits := mbits / duration

	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)
	fmt.Printf("%s%s[ÖZET]%s   %s%5.1f sn%s    %s%7.2f MByte%s   %s%s%7.2f Mbit/sn%s\n",
		Bold, Yellow, Reset,
		Yellow, duration, Reset,
		Green, mbytes, Reset,
		Bold, Cyan, avgMbits, Reset,
	)
}

func main() {
	serverMode := flag.Bool("s", false, "Sunucu modunda başlat (dinleyici)")
	clientIP := flag.String("c", "", "İstemci modunda başlat ve belirtilen IP'ye bağlan")

	flag.Usage = func() {
		fmt.Println("wperf - Ağ Hız Testi Aracı")
		fmt.Println("\nKullanım:")
		fmt.Println("  Sunucu olmak için: wperf -s")
		fmt.Println("  İstemci olmak için : wperf -c <sunucu_ip>")
	}

	flag.Parse()

	if *serverMode {
		runServer() // Sunucu kendisi içerisinde hata alana kadar loop yapar, her test bitiminde beklemeye döner
	} else if *clientIP != "" {
		for {
			runClient(*clientIP)
			fmt.Printf("\n%s[Test Durduruldu] Yeniden bağlanmak için ENTER tuşuna, programdan çıkmak için CTRL+C'ye basın...%s\n", Yellow, Reset)
			<-stdinLines
		}
	} else {
		// Hiç argüman verilmediyse her zaman interaktif menüde kalsın (çıkış için CTRL+C kullanabilirler)
		for {
			interactiveMenu()
			fmt.Println() // Menü yeniden başlatılmadan önce boşluk
		}
	}
}

func interactiveMenu() {
	fmt.Printf("%s===========================================================%s\n", Cyan, Reset)
	fmt.Printf("%s%s                 wperf - Ağ Hız Testi Aracı                %s\n", Bold, Green, Reset)
	fmt.Printf("%s===========================================================%s\n", Cyan, Reset)
	fmt.Println("Nasıl başlatmak istiyorsunuz?")
	fmt.Printf("  %s1)%s Sunucu (Gelen bağlantıları bekler)\n", Yellow, Reset)
	fmt.Printf("  %s2)%s İstemci (Karşı cihaza veri gönderir)\n", Yellow, Reset)
	fmt.Print("\nSeçiminiz (1 veya 2): ")

	choice := <-stdinLines
	choice = strings.TrimSpace(choice)

	if choice == "1" {
		runServer()
	} else if choice == "2" {
		fmt.Print("\nBağlanılacak Sunucunun IP Adresini Girin (Örn: 192.168.1.10): ")
		ip := <-stdinLines
		ip = strings.TrimSpace(ip)

		if ip != "" {
			runClient(ip)
			fmt.Printf("\n%s[Test sonlandı] Ana menüye dönmek için ENTER tuşuna, çıkmak için CTRL+C'ye basın...%s\n", Yellow, Reset)
			<-stdinLines
		} else {
			fmt.Printf("%sHata: Geçersiz IP adresi girdiniz.%s\n", Red, Reset)
			fmt.Printf("\n%sMenüye dönmek için ENTER tuşuna basın...%s\n", Yellow, Reset)
			<-stdinLines
		}
	} else {
		fmt.Printf("%sHata: Geçersiz seçim yaptınız.%s\n", Red, Reset)
		fmt.Printf("\n%sMenüye dönmek için ENTER tuşuna basın...%s\n", Yellow, Reset)
		<-stdinLines
	}
}

func runServer() {
	addr := ":" + defaultPort
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("%sSunucu başlatılamadı: %v%s\n", Red, err, Reset)
		fmt.Printf("\n%sMenüye dönmek için ENTER tuşuna basın...%s\n", Yellow, Reset)
		<-stdinLines
		return
	}
	defer listener.Close()

	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)
	fmt.Printf("%swperf sunucusu dinleniyor... Port: %s%s\n", Green, defaultPort, Reset)
	
	ips := getLocalIPs()
	if len(ips) > 0 {
		fmt.Printf("%s\nBu makinenin yerel IP Adresleri (İstemciye bunu girebilirsiniz):%s\n", Yellow, Reset)
		for _, ip := range ips {
			fmt.Printf("   -> %s%s%s%s\n", Bold, Green, ip, Reset)
		}
	}

	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)
	fmt.Printf("%s(İstemci bağlandığında test başlar, sunucuyu kapatmak için CTRL+C yapabilirsiniz)%s\n", Yellow, Reset)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Bağlantı kabul edilemedi: %v\n", err)
			continue
		}
		// Sadece bir bağlantıyı sırayla işler (Test süresince diğerlerini bekletir)
		// Bağlantı bitip özet yazıldıktan sonra döngü tekrar eder ve başlangıç durumuna (listener) döner.
		handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("\n%s[+] Yeni test bağlantısı kabul edildi: %s%s\n", Green, conn.RemoteAddr(), Reset)

	var totalBytesRecv uint64
	atomic.StoreInt32(&stopTest, 0)
	startTime := time.Now()

	// Hız ölçüm goroutine'i
	go func() {
		var lastBytes uint64
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		printHeader()
		for range ticker.C {
			if atomic.LoadInt32(&stopTest) == 1 {
				break
			}
			currentBytes := atomic.LoadUint64(&totalBytesRecv)
			bytesInSec := currentBytes - lastBytes
			lastBytes = currentBytes

			mbytes := float64(bytesInSec) / (1024 * 1024)
			mbits := (float64(bytesInSec) * 8) / (1000 * 1000)

			fmt.Printf("1 sn               %s%7.2f MByte%s     %s%s%7.2f Mbit/sn%s\n", Green, mbytes, Reset, Bold, Cyan, mbits, Reset)
		}
	}()

	buf := make([]byte, bufferSize)
	for {
		n, err := conn.Read(buf)
		if n > 0 {
			atomic.AddUint64(&totalBytesRecv, uint64(n))
		}
		if err != nil {
			if err != io.EOF && atomic.LoadInt32(&stopTest) == 0 && err.Error() != "use of closed network connection" {
				fmt.Printf("\n%sOku hatası veya İstemci Kapatıldı: %v%s\n", Red, err, Reset)
			}
			break
		}
	}

	atomic.StoreInt32(&stopTest, 1) // Hız okuyucuyu kes
	time.Sleep(50 * time.Millisecond) // Çıktı karışmasını önle
	
	printSummary(atomic.LoadUint64(&totalBytesRecv), startTime)
	fmt.Printf("%s[-] İstemci bağını kesti, sunucu yeni bağlantılara hazır.%s\n", Yellow, Reset)
}

func runClient(ip string) {
	addr := ip + ":" + defaultPort
	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)
	fmt.Printf("%swperf istemcisi sunucuya bağlanıyor... Hedef: %s%s\n", Green, addr, Reset)
	fmt.Printf("%s-----------------------------------------------------------%s\n", Cyan, Reset)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Printf("%sSunucuya bağlanılamadı: %v%s\n", Red, err, Reset)
		return
	}
	defer conn.Close()

	fmt.Printf("%s[+] Bağlantı başarılı: %s -> %s%s\n", Green, conn.LocalAddr(), conn.RemoteAddr(), Reset)
	fmt.Printf("%s(Test başladı. Durdurup Özeti görmek ve menüye dönmek için ENTER tuşuna basın)%s\n", Bold, Reset)

	var totalBytesSent uint64
	atomic.StoreInt32(&stopTest, 0)
	startTime := time.Now()

	done := make(chan bool)

	// Enter tuşu ile testi sadece durduran ve özete düşmesini sağlayan dinleyici
	go func() {
		select {
		case <-stdinLines: // ENTER tuşuna basıldı
			atomic.StoreInt32(&stopTest, 1)
			conn.Close() // Bağlantı koptuğu an Write fonksiyonu hata verip for döngüsünü bitirir.
		case <-done: // Eğer sunucu koparsa, bu goroutine kendiliğinden kapanır ve ENTER isteğini çalmaz
		}
	}()

	// Hız ölçüm goroutine'i
	go func() {
		var lastBytes uint64
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		printHeader()
		for range ticker.C {
			if atomic.LoadInt32(&stopTest) == 1 {
				break
			}
			currentBytes := atomic.LoadUint64(&totalBytesSent)
			bytesInSec := currentBytes - lastBytes
			lastBytes = currentBytes

			mbytes := float64(bytesInSec) / (1024 * 1024)
			mbits := (float64(bytesInSec) * 8) / (1000 * 1000)

			fmt.Printf("1 sn               %s%7.2f MByte%s     %s%s%7.2f Mbit/sn%s\n", Green, mbytes, Reset, Bold, Cyan, mbits, Reset)
		}
	}()

	buf := make([]byte, bufferSize)
	for i := 0; i < bufferSize; i++ {
		buf[i] = byte(rand.Intn(256))
	}

	for {
		if atomic.LoadInt32(&stopTest) == 1 {
			break
		}
		n, err := conn.Write(buf)
		if n > 0 {
			atomic.AddUint64(&totalBytesSent, uint64(n))
		}
		if err != nil {
			break
		}
	}

	atomic.StoreInt32(&stopTest, 1)
	close(done) // ENTER dinleyicisini bitir ki sonraki input'ları çalmasın
	time.Sleep(50 * time.Millisecond)
	
	printSummary(atomic.LoadUint64(&totalBytesSent), startTime)
}

func getLocalIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// Sadece IPv4 bekle ve localhost (127.0.0.1) olmayanları al
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				ips = append(ips, ip.String())
			}
		}
	}
	return ips
}
