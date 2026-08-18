package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

var asamalar = map[string]int{
	"bekliyor":       0,
	"hesap açıldı":   1,
	"dışa aktarıldı": 2,
	"taşındı":        3,
	"içe aktarıldı":  4,
	"tamamlandı":     5,
}

func durumSira(durum string) int {
	if sira, varMi := asamalar[durum]; varMi {
		return sira
	}
	return 0
}

type adim struct {
	hedef    string
	ad       string
	calistir func(*Ayar, *Hesap) error
}

func adimlar() []adim {
	return []adim{
		{"hesap açıldı", "hesap açma", adimHesapAc},
		{"dışa aktarıldı", "dışa aktarım", adimDisaAktar},
		{"taşındı", "taşıma", adimTasi},
		{"içe aktarıldı", "içe aktarım", adimIceAktar},
	}
}

func hesabiIsle(ayar *Ayar, depo *Depo, h *Hesap) {
	for _, a := range adimlar() {
		if durumSira(h.Durum) >= durumSira(a.hedef) {
			continue
		}
		logla("...   %-32s %s başlıyor", h.Eposta, a.ad)
		if err := a.calistir(ayar, h); err != nil {
			depo.guncelle(h.Eposta, func(x *Hesap) {
				x.Basarisiz = true
				x.BasarisizAdim = a.ad
				x.Hata = err.Error()
			})
			logla("HATA  %-32s [%s] %v", h.Eposta, a.ad, err)
			return
		}
		depo.guncelle(h.Eposta, func(x *Hesap) {
			x.Durum = a.hedef
			x.Basarisiz = false
			x.BasarisizAdim = ""
			x.Hata = ""
		})
		logla("OK    %-32s -> %s", h.Eposta, a.hedef)
	}
	if ayar.Temizlik {
		adimTemizle(ayar, h)
	}
	depo.guncelle(h.Eposta, func(x *Hesap) { x.Durum = "tamamlandı" })
	logla("BİTTİ %s", h.Eposta)
}

func adimHesapAc(ayar *Ayar, h *Hesap) error {
	p := ayar.Carbonio.ProvKomut
	var b strings.Builder
	fmt.Fprintf(&b, "set -e\n")
	fmt.Fprintf(&b, "%s gd %s >/dev/null 2>&1 || %s cd %s zimbraAuthMech zimbra\n",
		p, tirnakla(h.AlanAdi), p, tirnakla(h.AlanAdi))

	nitelikler := nitelikCiftleri(
		"cn", ilkDolu(h.GorunenAd, h.Ad, yerelKisim(h.Eposta)),
		"displayName", h.GorunenAd,
		"givenName", h.Ad,
		"sn", h.Soyad,
	)
	parola := ayar.VarsayilanSifre
	if parola == "" {
		parola = rastgeleParola()
	}
	if h.SifreOzeti == "" {
		logla("uyarı: %s için şifre özeti okunamadı; hesap geçici parolayla açılacak, kullanıcıya yeni parola bildirilmeli", h.Eposta)
	}
	fmt.Fprintf(&b, "if ! %s ga %s >/dev/null 2>&1; then %s ca %s %s%s; fi\n",
		p, tirnakla(h.Eposta), p, tirnakla(h.Eposta), tirnakla(parola), nitelikler)

	if h.SifreOzeti != "" {
		fmt.Fprintf(&b, "%s ma %s userPassword %s\n", p, tirnakla(h.Eposta), tirnakla(h.SifreOzeti))
	}
	if ayar.KotaUygula && h.Kota != "" {
		fmt.Fprintf(&b, "%s ma %s zimbraMailQuota %s\n", p, tirnakla(h.Eposta), tirnakla(h.Kota))
	}
	if hedefCos := ayar.CosEsleme[h.Cos]; hedefCos != "" {
		fmt.Fprintf(&b, "%s ma %s zimbraCOSId %s\n", p, tirnakla(h.Eposta), tirnakla(hedefCos))
	}
	cikti, hataCikti, err := sshCalistir(ayar.Carbonio, b.String())
	if err != nil {
		return fmt.Errorf("%s", ilkDolu(hataCikti, cikti, err.Error()))
	}
	return nil
}

func adimDisaAktar(ayar *Ayar, h *Hesap) error {
	m := ayar.Zimbra.PostaKomut
	dosya := uzakTgz(ayar, h.Eposta)
	betik := fmt.Sprintf("set -e\nmkdir -p %s\n%s -z -m %s -t 0 getRestURL '/?fmt=tgz' > %s\nstat -c %%s %s\n",
		tirnakla(ayar.UzakCalismaDizini), m, tirnakla(h.Eposta), tirnakla(dosya), tirnakla(dosya))
	cikti, hataCikti, err := sshCalistir(ayar.Zimbra, betik)
	if err != nil {
		return fmt.Errorf("%s", ilkDolu(hataCikti, cikti, err.Error()))
	}
	if bayt, hata := strconv.ParseInt(strings.TrimSpace(sonSatir(cikti)), 10, 64); hata == nil {
		h.KaynakBayt = bayt
	}
	return nil
}

func adimTasi(ayar *Ayar, h *Hesap) error {
	dosya := uzakTgz(ayar, h.Eposta)
	switch ayar.AktarimModu {
	case "dogrudan":
		tazeCikti, _, tazeHata := sshCalistir(ayar.Zimbra, "stat -c %s "+tirnakla(dosya))
		if tazeHata == nil {
			if bayt, hata := strconv.ParseInt(strings.TrimSpace(sonSatir(tazeCikti)), 10, 64); hata == nil {
				h.KaynakBayt = bayt
			}
		}
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
		betik := fmt.Sprintf("set -e\nmkdir -p %s\nscp -P %d %s-o BatchMode=yes -o StrictHostKeyChecking=accept-new %s@%s:%s %s\nstat -c %%s %s\n",
			tirnakla(ayar.UzakCalismaDizini), zPort, anahtar,
			tirnakla(zKullanici), tirnakla(d.ZimbraAdres), tirnakla(dosya),
			tirnakla(ayar.UzakCalismaDizini+"/"), tirnakla(dosya))
		cikti, hataCikti, err := sshCalistir(ayar.Carbonio, betik)
		if err != nil {
			return fmt.Errorf("%s", ilkDolu(hataCikti, cikti, err.Error()))
		}
		if bayt, hata := strconv.ParseInt(strings.TrimSpace(sonSatir(cikti)), 10, 64); hata == nil {
			h.HedefBayt = bayt
		}
	case "aktarmali":
		if err := os.MkdirAll(ayar.AktarmaliYerelDizin, 0o700); err != nil {
			return err
		}
		yerel := filepath.Join(ayar.AktarmaliYerelDizin, temizAd(h.Eposta)+".tgz")
		if err := scpIndir(ayar.Zimbra, dosya, yerel); err != nil {
			return fmt.Errorf("indirme: %w", err)
		}
		if bilgi, hata := os.Stat(yerel); hata == nil {
			h.KaynakBayt = bilgi.Size()
		}
		if _, hataCikti, err := sshCalistir(ayar.Carbonio, "mkdir -p "+tirnakla(ayar.UzakCalismaDizini)); err != nil {
			return fmt.Errorf("hedef dizin: %s", ilkDolu(hataCikti, err.Error()))
		}
		if err := scpYukle(ayar.Carbonio, yerel, dosya); err != nil {
			return fmt.Errorf("yükleme: %w", err)
		}
		cikti, _, err := sshCalistir(ayar.Carbonio, "stat -c %s "+tirnakla(dosya))
		if err == nil {
			if bayt, hata := strconv.ParseInt(strings.TrimSpace(sonSatir(cikti)), 10, 64); hata == nil {
				h.HedefBayt = bayt
			}
		}
		_ = os.Remove(yerel)
	default:
		return fmt.Errorf("bilinmeyen aktarimModu: %s", ayar.AktarimModu)
	}
	if h.KaynakBayt > 0 && h.HedefBayt > 0 && h.KaynakBayt != h.HedefBayt {
		return fmt.Errorf("boyut uyuşmuyor: kaynak=%d hedef=%d", h.KaynakBayt, h.HedefBayt)
	}
	return nil
}

func adimIceAktar(ayar *Ayar, h *Hesap) error {
	m := ayar.Carbonio.PostaKomut
	dosya := uzakTgz(ayar, h.Eposta)
	betik := fmt.Sprintf("set -e\n%s -z -m %s -t 0 postRestURL '/?fmt=tgz&resolve=skip' %s\n",
		m, tirnakla(h.Eposta), tirnakla(dosya))
	cikti, hataCikti, err := sshCalistir(ayar.Carbonio, betik)
	if err != nil {
		return fmt.Errorf("%s", ilkDolu(hataCikti, cikti, err.Error()))
	}
	return nil
}

func adimTemizle(ayar *Ayar, h *Hesap) {
	dosya := uzakTgz(ayar, h.Eposta)
	_, _, _ = sshCalistir(ayar.Zimbra, "rm -f "+tirnakla(dosya))
	_, _, _ = sshCalistir(ayar.Carbonio, "rm -f "+tirnakla(dosya))
}

func carbonioHazirla(ayar *Ayar) error {
	betik := "zmlocalconfig -e socket_so_timeout=3000000 >/dev/null 2>&1 || true\nzmlocalconfig --reload >/dev/null 2>&1 || true\nmkdir -p " + tirnakla(ayar.UzakCalismaDizini) + "\necho ok\n"
	_, hataCikti, err := sshCalistir(ayar.Carbonio, betik)
	if err != nil {
		return fmt.Errorf("%s", ilkDolu(hataCikti, err.Error()))
	}
	return nil
}

func tasimayiCalistir(ayar *Ayar, depo *Depo, yeniden bool, sadece []string) {
	if err := carbonioHazirla(ayar); err != nil {
		logla("uyarı: Carbonio hazırlık: %v", err)
	}
	kanal := make(chan struct{}, ayar.EszamanliSayi)
	var bekle sync.WaitGroup
	islenen := 0
	for _, h := range depo.listele() {
		if h.Durum == "tamamlandı" {
			continue
		}
		if len(sadece) > 0 && !icerir(sadece, h.Eposta) {
			continue
		}
		if h.Basarisiz && !yeniden {
			continue
		}
		if h.Basarisiz && yeniden {
			depo.guncelle(h.Eposta, func(x *Hesap) { x.Basarisiz = false; x.Hata = ""; x.BasarisizAdim = "" })
		}
		islenen++
		bekle.Add(1)
		kanal <- struct{}{}
		go func(hesap *Hesap) {
			defer bekle.Done()
			defer func() { <-kanal }()
			hesabiIsle(ayar, depo, hesap)
		}(h)
	}
	bekle.Wait()
	logla("taşıma turu bitti: %d hesap işlendi", islenen)
}

func sadeceHesapAc(ayar *Ayar, depo *Depo, sadece []string) {
	kanal := make(chan struct{}, ayar.EszamanliSayi)
	var bekle sync.WaitGroup
	for _, h := range depo.listele() {
		if durumSira(h.Durum) >= durumSira("hesap açıldı") {
			continue
		}
		if len(sadece) > 0 && !icerir(sadece, h.Eposta) {
			continue
		}
		bekle.Add(1)
		kanal <- struct{}{}
		go func(hesap *Hesap) {
			defer bekle.Done()
			defer func() { <-kanal }()
			if err := adimHesapAc(ayar, hesap); err != nil {
				depo.guncelle(hesap.Eposta, func(x *Hesap) {
					x.Basarisiz = true
					x.BasarisizAdim = "hesap açma"
					x.Hata = err.Error()
				})
				logla("HATA  %-32s [hesap açma] %v", hesap.Eposta, err)
				return
			}
			depo.guncelle(hesap.Eposta, func(x *Hesap) { x.Durum = "hesap açıldı"; x.Basarisiz = false })
			logla("OK    %-32s -> hesap açıldı", hesap.Eposta)
		}(h)
	}
	bekle.Wait()
}
