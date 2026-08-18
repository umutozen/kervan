package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

const afis = `
 _  __ _____  ____  __     __    _     _   _
| |/ /| ____||  _ \ \ \   / /   / \   | \ | |
| ' / |  _|  | |_) | \ \ / /   / _ \  |  \| |
| . \ | |___ |  _ <   \ V /   / ___ \ | |\  |
|_|\_\|_____||_| \_\   \_/   /_/   \_\|_| \_|

     Göçtü kervan kaldık dağlar başında..
`

const kullanimMetni = `kervan - Zimbra'dan Carbonio'ya hesap ve posta kutusu taşıma aracı

Kullanım:
  kervan <komut> [bayraklar]

Komutlar:
  kontrol     Başlamadan önce her şey yerinde mi diye denetle
  envanter    Zimbra'dan alan adı + hesap + meta + şifre özeti envanterini çıkar
  olustur     Sadece hedefte hesapları oluştur (veri taşımadan)
  tasi        Hesapları uçtan uca taşı (hesap açma > dışa aktarım > taşıma > içe aktarım)
  esitle      Taşınan hesaplara sonradan gelen postaları eşitle (imapsync)
  durum       Durum tablosunu göster ve rapor.csv yaz
  dogrula     Kaynak/hedef tgz boyutlarını karşılaştır

Ortak bayraklar:
  -ayar <yol>     ayar dosyası (varsayılan ayar.json)
  -sadece <a,b>   sadece bu e-postalar
  -yeniden        başarısız hesapları yeniden dene (tasi)
  -prova          komutları çalıştırma, sadece yazdır

Örnek:
  kervan envanter
  kervan tasi -sadece ali@x.gov.tr,veli@x.gov.tr
  kervan tasi -yeniden
`

func main() {
	fmt.Print(afis)
	if len(os.Args) < 2 {
		fmt.Print(kullanimMetni)
		os.Exit(2)
	}
	komut := os.Args[1]
	bayraklar := flag.NewFlagSet(komut, flag.ExitOnError)
	ayarYolu := bayraklar.String("ayar", "ayar.json", "ayar dosyası")
	sadece := bayraklar.String("sadece", "", "sadece bu e-postalar (virgülle)")
	yeniden := bayraklar.Bool("yeniden", false, "başarısızları yeniden dene")
	prova := bayraklar.Bool("prova", false, "komutları çalıştırma, yazdır")
	_ = bayraklar.Parse(os.Args[2:])

	if komut == "yardim" || komut == "-h" || komut == "--help" {
		fmt.Print(kullanimMetni)
		return
	}

	ayar, err := ayarYukle(*ayarYolu)
	denetle(err)
	gProva = *prova
	depo, err := depoYukle(ayar.DurumDosyasi)
	denetle(err)
	sadeceListe := virgulAyir(*sadece)

	switch komut {
	case "kontrol":
		kontrolEt(ayar)
	case "esitle":
		esitle(ayar, depo, sadeceListe)
	case "envanter":
		denetle(envanterCikar(ayar, depo))
	case "olustur":
		sadeceHesapAc(ayar, depo, sadeceListe)
	case "tasi":
		tasimayiCalistir(ayar, depo, *yeniden, sadeceListe)
	case "durum":
		durumYazdir(depo)
	case "dogrula":
		dogrula(depo, sadeceListe)
	default:
		fmt.Printf("bilinmeyen komut: %s\n\n", komut)
		fmt.Print(kullanimMetni)
		os.Exit(2)
	}
}

func durumYazdir(depo *Depo) {
	sayilar := map[string]int{}
	var basarisizlar []*Hesap
	for _, h := range depo.listele() {
		anahtar := h.Durum
		if h.Basarisiz {
			anahtar = "başarısız"
		}
		sayilar[anahtar]++
		if h.Basarisiz {
			basarisizlar = append(basarisizlar, h)
		}
	}
	anahtarlar := make([]string, 0, len(sayilar))
	for a := range sayilar {
		anahtarlar = append(anahtarlar, a)
	}
	sort.Strings(anahtarlar)
	fmt.Println("Durum özeti:")
	for _, a := range anahtarlar {
		fmt.Printf("  %-16s %d\n", a, sayilar[a])
	}
	fmt.Printf("  %-16s %d\n", "TOPLAM", len(depo.Hesaplar))

	if len(basarisizlar) > 0 {
		fmt.Println("\nBaşarısız hesaplar:")
		for _, h := range basarisizlar {
			fmt.Printf("  %-32s [%s] %s\n", h.Eposta, h.BasarisizAdim, tekSatir(h.Hata))
		}
	}
	raporYaz(depo, "rapor.csv")
	fmt.Println("\nrapor yazıldı: rapor.csv")
}

func raporYaz(depo *Depo, yol string) {
	var b strings.Builder
	b.WriteString("eposta,alanAdi,durum,basarisiz,basarisizAdim,kaynakBayt,hedefBayt,guncelleme,hata\n")
	for _, h := range depo.listele() {
		b.WriteString(strings.Join([]string{
			csvAlan(h.Eposta), csvAlan(h.AlanAdi), csvAlan(h.Durum), fmt.Sprintf("%t", h.Basarisiz),
			csvAlan(h.BasarisizAdim), fmt.Sprintf("%d", h.KaynakBayt), fmt.Sprintf("%d", h.HedefBayt),
			csvAlan(h.Guncelleme), csvAlan(tekSatir(h.Hata)),
		}, ","))
		b.WriteString("\n")
	}
	_ = os.WriteFile(yol, []byte(b.String()), 0o600)
}

func dogrula(depo *Depo, sadece []string) {
	tamam, uyari := 0, 0
	for _, h := range depo.listele() {
		if len(sadece) > 0 && !icerir(sadece, h.Eposta) {
			continue
		}
		if durumSira(h.Durum) < durumSira("taşındı") {
			continue
		}
		if h.KaynakBayt > 0 && h.KaynakBayt == h.HedefBayt {
			tamam++
			continue
		}
		uyari++
		fmt.Printf("UYARI %-32s kaynak=%d hedef=%d\n", h.Eposta, h.KaynakBayt, h.HedefBayt)
	}
	fmt.Printf("doğrulama: %d tamam, %d uyarı\n", tamam, uyari)
	fmt.Println("not: bu boyut kontrolüdür. Mesaj sayısı mutabakatı için her iki sunucuda")
	fmt.Println("     `zmmailbox -z -m <hesap> getFolder /` çıktısını karşılaştırın.")
}

func csvAlan(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
	}
	return s
}

func tekSatir(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\n", " "), "\r", " "))
}
