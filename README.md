# wperf
wperf, sistemler arası ağ bant genişliğini ve hızını ölçmek için kullanabileceğiniz `iperf` benzeri, kullanımı kolay ve taşınabilir bir Ağ Hız Testi (Network Speed Test) aracıdır. Go (Golang) ile geliştirilmiştir, hafiftir ve bağımlılık gerektirmez.

## Özellikler
* **Çapraz Platform:** Windows, Linux, macOS ve Android (Termux) ile tam uyumlu.
* **Kuruluma Gerek Yok:** Tek bir çalıştırılabilir dosya olarak çalışır.
* **Kolay Kullanım:** Sadece `-s` (sunucu) ve `-c` (istemci) parametreleri ile anında kullanıma hazır.
* **Düşük Gecikme & Yüksek Performans:** Özel ayarlanan kilitler ve tampon(buffer) mekanizması ile diske yazma işlemi olmaksızın en doğru donanım hızını ölçer.

## Kurulum
Yüklü bir **Go** ortamınız varsa, komut satırı üzerinden tek bir komutla anında kurabilirsiniz:

```bash
go install github.com/KULLANICI_ADINIZ/wperf@latest
```
*Not: Yukarıdaki adresi GitHub hesabınıza koyduktan sonra kendi adresinizle değiştirin.*

### Ya da Hazır Olarak İndirin (Go Yüklemeden)
GitHub'da **Releases** sekmesi altında yer alan Windows (.exe) veya Linux/Mac binary dosyalarından sisteminize uygun olanı indirebilirsiniz.

Android üzerinde çalıştırmak için **Termux** kullanıyorsanız:
```bash
pkg update && pkg install golang
go install github.com/KULLANICI_ADINIZ/wperf@latest
```

## Nasıl Kullanılır?

Testi gerçekleştirmek için iki adet cihaza ihtiyacınız vardır. (Test amaçlı kendi cihazınızda da yapabilirsiniz)

### 1. Sunucuyu Başlatın (Dinleyici Kısım)
Bir cihazı hız testi sunucusu olarak ayarlamak için terminal(veya CMD)'de şu komutu çalıştırın:
```bash
wperf -s
```
Bu komut arka planda otomatik olarak `:5202` portunu dinlemeye başlayacaktır.

### 2. İstemciyi Başlatın (Gönderici Kısım)
Diğer cihazda, sunucunun yerel(local) veya dış(public) IP adresine bağlanın:
```bash
wperf -c 192.168.1.10
```
Bağlantı kurulduğu an bant genişliği testi başlar ve saniyelik bazda ne kadar hızda veri aktarıldığı terminal ekranında Mbyte ve Mbit/sn cinsinden canlı olarak gösterilir. Testi bitirmek için `CTRL+C` yapabilirsiniz.

## İletişim / Katkıda Bulunma
Herhangi bir sorun bulduğunuzda *Issue* açmaktan veya *Pull Request* göndermekten çekinmeyin!
