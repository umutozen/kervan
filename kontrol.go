package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func kontrolEt(ayar *Ayar) {
	fmt.Println("Taşıma öncesi sağlık kontrolü:")
	fmt.Println()
	sorun := 0

	fmt.Println("Bu makine:")
	for _, arac := range []string{"ssh", "scp"} {
		if _, err := exec.LookPath(arac); err != nil {
			fmt.Printf("  %-20s HATA: bulunamadı\n", arac)
			sorun++
		} else {
			fmt.Printf("  %-20s tamam\n", arac)
		}
	}

	fmt.Println()
	sorun += sunucuKontrol("Zimbra", ayar.Zimbra, ayar.UzakCalismaDizini)
	fmt.Println()
	sorun += sunucuKontrol("Carbonio", ayar.Carbonio, ayar.UzakCalismaDizini)

	if ayar.AktarimModu == "dogrudan" {
		fmt.Println()
		fmt.Println("Doğrudan aktarım (Carbonio'dan Zimbra'ya ssh):")
		d := ayar.Dogrudan
		zKullanici := ilkDolu(d.ZimbraKullanici, ayar.Zimbra.Kullanici)
		zPort := d.ZimbraPort
		if zPort == 0 {
			zPort = ayar.Zimbra.Port
		}
		anahtar := ""
		if d.CarbonioAnahtar != "" {
			anahtar = "-i " + tirnakla(d.CarbonioAnahtar) + " "
		}
		betik := fmt.Sprintf("ssh -p %d %s-o BatchMode=yes -o ConnectTimeout=10 -o StrictHostKeyChecking=accept-new %s@%s 'echo dogrudan-ok' 2>/dev/null || echo dogrudan-hata",
			zPort, anahtar, zKullanici, d.ZimbraAdres)
		cikti, _, err := sshCalistir(ayar.Carbonio, betik)
		if err == nil && strings.Contains(cikti, "dogrudan-ok") {
			fmt.Printf("  %-20s tamam\n", "bağlantı")
		} else {
			fmt.Printf("  %-20s HATA: Carbonio üzerindeki anahtarla Zimbra'ya girilemiyor\n", "bağlantı")
			sorun++
		}
	}

	fmt.Println()
	if sorun == 0 {
		fmt.Println("Her şey hazır, taşımaya başlayabilirsiniz.")
	} else {
		fmt.Printf("%d sorun bulundu. Yukarıdaki HATA satırlarını giderdikten sonra tekrar deneyin.\n", sorun)
	}
}

func sunucuKontrol(rol string, s SunucuAyar, dizin string) int {
	fmt.Printf("%s (%s@%s:%d):\n", rol, s.Kullanici, s.Adres, s.Port)
	sorun := 0
	provArac := strings.Fields(s.ProvKomut)[0]
	kutuArac := strings.Fields(s.PostaKomut)[0]

	var b strings.Builder
	fmt.Fprintf(&b, "echo K:baglanti:ok\n")
	fmt.Fprintf(&b, "command -v %s >/dev/null 2>&1 && echo K:prov:ok || echo K:prov:yok\n", provArac)
	fmt.Fprintf(&b, "command -v %s >/dev/null 2>&1 && echo K:kutu:ok || echo K:kutu:yok\n", kutuArac)
	fmt.Fprintf(&b, "mkdir -p %s >/dev/null 2>&1 && touch %s/.kervan_test >/dev/null 2>&1 && rm -f %s/.kervan_test && echo K:dizin:ok || echo K:dizin:hata\n",
		tirnakla(dizin), tirnakla(dizin), tirnakla(dizin))
	fmt.Fprintf(&b, "df -Pk %s 2>/dev/null | awk 'NR==2{print \"K:disk:\"$4}'\n", tirnakla(dizin))

	cikti, hataCikti, err := sshCalistir(s, b.String())
	if err != nil {
		fmt.Printf("  %-20s HATA: bağlanılamadı: %s\n", "bağlantı", tekSatir(ilkDolu(hataCikti, err.Error())))
		return 1
	}
	deger := map[string]string{}
	for _, satir := range strings.Split(cikti, "\n") {
		if strings.HasPrefix(satir, "K:") {
			parca := strings.SplitN(satir, ":", 3)
			if len(parca) == 3 {
				deger[parca[1]] = parca[2]
			}
		}
	}

	fmt.Printf("  %-20s tamam\n", "bağlantı")
	if deger["prov"] == "ok" {
		fmt.Printf("  %-20s tamam\n", provArac)
	} else {
		fmt.Printf("  %-20s HATA: bulunamadı (doğru kullanıcıyla mı bağlanıyorsunuz?)\n", provArac)
		sorun++
	}
	if deger["kutu"] == "ok" {
		fmt.Printf("  %-20s tamam\n", kutuArac)
	} else {
		fmt.Printf("  %-20s HATA: bulunamadı (doğru kullanıcıyla mı bağlanıyorsunuz?)\n", kutuArac)
		sorun++
	}
	if deger["dizin"] == "ok" {
		fmt.Printf("  %-20s yazılabilir (%s)\n", "çalışma dizini", dizin)
	} else {
		fmt.Printf("  %-20s HATA: %s dizinine yazılamıyor\n", "çalışma dizini", dizin)
		sorun++
	}
	if kb, hata := strconv.ParseInt(deger["disk"], 10, 64); hata == nil {
		gb := float64(kb) / 1024 / 1024
		not := "tamam"
		if gb < 5 {
			not = "DİKKAT: az, büyük posta kutularında yetmeyebilir"
		}
		fmt.Printf("  %-20s %.1f GB boş (%s)\n", "disk", gb, not)
	}
	return sorun
}
