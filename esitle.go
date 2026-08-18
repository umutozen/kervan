package main

import (
	"fmt"
	"strings"
	"sync"
)

func esitle(ayar *Ayar, depo *Depo, sadece []string) {
	e := ayar.Esitleme
	if e.ZimbraYonetici == "" || e.ZimbraYoneticiSifre == "" || e.CarbonioYonetici == "" || e.CarbonioYoneticiSifre == "" {
		denetle(fmt.Errorf("ayar dosyasındaki 'esitleme' bölümünü doldurun: iki sunucunun IMAP yönetici hesabı ve şifresi gerekli"))
	}
	sunucu := ayar.Carbonio
	if e.CalistigiYer == "zimbra" {
		sunucu = ayar.Zimbra
	}
	if !gProva {
		cikti, _, err := sshCalistir(sunucu, "command -v imapsync >/dev/null 2>&1 && echo var || echo yok")
		if err != nil {
			denetle(fmt.Errorf("eşitleme sunucusuna (%s) bağlanılamadı: %v", sunucu.Adres, err))
		}
		if strings.TrimSpace(sonSatir(cikti)) != "var" {
			denetle(fmt.Errorf("imapsync %s üzerinde kurulu değil. Kurulum: apt install imapsync (Debian/Ubuntu) ya da dnf install imapsync (RHEL/Rocky)", sunucu.Adres))
		}
	}

	kanal := make(chan struct{}, ayar.EszamanliSayi)
	var bekle sync.WaitGroup
	islenen := 0
	for _, h := range depo.listele() {
		if durumSira(h.Durum) < durumSira("içe aktarıldı") {
			continue
		}
		if len(sadece) > 0 && !icerir(sadece, h.Eposta) {
			continue
		}
		islenen++
		bekle.Add(1)
		kanal <- struct{}{}
		go func(hesap *Hesap) {
			defer bekle.Done()
			defer func() { <-kanal }()
			logla("...   %-32s eşitleme başlıyor", hesap.Eposta)
			ozet, err := hesapEsitle(ayar, sunucu, hesap)
			if err != nil {
				logla("HATA  %-32s [eşitleme] %v", hesap.Eposta, err)
				return
			}
			depo.guncelle(hesap.Eposta, func(x *Hesap) { x.SonEsitleme = simdi() })
			if ozet != "" {
				logla("OK    %-32s eşitlendi (%s)", hesap.Eposta, ozet)
			} else {
				logla("OK    %-32s eşitlendi", hesap.Eposta)
			}
		}(h)
	}
	bekle.Wait()
	if islenen == 0 {
		logla("eşitlenecek hesap yok: önce 'kervan tasi' ile hesapları taşıyın")
		return
	}
	logla("eşitleme bitti: %d hesap", islenen)
}

func hesapEsitle(ayar *Ayar, sunucu SunucuAyar, h *Hesap) (string, error) {
	e := ayar.Esitleme
	zAdres := ilkDolu(e.ZimbraImapAdres, ayar.Zimbra.Adres)
	cAdres := e.CarbonioImapAdres
	if cAdres == "" {
		if e.CalistigiYer == "zimbra" {
			cAdres = ayar.Carbonio.Adres
		} else {
			cAdres = "127.0.0.1"
		}
	}
	zPort := e.ZimbraImapPort
	if zPort == 0 {
		zPort = 993
	}
	cPort := e.CarbonioImapPort
	if cPort == 0 {
		cPort = 993
	}
	dizin := strings.TrimRight(ayar.UzakCalismaDizini, "/")
	ad := temizAd(h.Eposta)
	sifre1 := fmt.Sprintf("%s/.e1_%s", dizin, ad)
	sifre2 := fmt.Sprintf("%s/.e2_%s", dizin, ad)
	ek := ""
	if len(e.EkSecenekler) > 0 {
		ek = " " + strings.Join(e.EkSecenekler, " ")
	}

	var b strings.Builder
	fmt.Fprintf(&b, "umask 077\nmkdir -p %s\n", tirnakla(dizin))
	fmt.Fprintf(&b, "printf %%s %s > %s\n", tirnakla(e.ZimbraYoneticiSifre), tirnakla(sifre1))
	fmt.Fprintf(&b, "printf %%s %s > %s\n", tirnakla(e.CarbonioYoneticiSifre), tirnakla(sifre2))
	fmt.Fprintf(&b, "imapsync --host1 %s --port1 %d --ssl1 --user1 %s --authuser1 %s --passfile1 %s --authmech1 PLAIN --host2 %s --port2 %d --ssl2 --user2 %s --authuser2 %s --passfile2 %s --authmech2 PLAIN --nolog%s\n",
		tirnakla(zAdres), zPort, tirnakla(h.Eposta), tirnakla(e.ZimbraYonetici), tirnakla(sifre1),
		tirnakla(cAdres), cPort, tirnakla(h.Eposta), tirnakla(e.CarbonioYonetici), tirnakla(sifre2), ek)
	fmt.Fprintf(&b, "rc=$?\nrm -f %s %s\nexit $rc\n", tirnakla(sifre1), tirnakla(sifre2))

	cikti, hataCikti, err := sshCalistir(sunucu, b.String())
	if err != nil {
		return "", fmt.Errorf("%s", tekSatir(ilkDolu(hataCikti, sonSatir(cikti), err.Error())))
	}
	for _, satir := range strings.Split(cikti, "\n") {
		if strings.Contains(strings.ToLower(satir), "messages transferred") {
			return strings.TrimSpace(satir), nil
		}
	}
	return "", nil
}
