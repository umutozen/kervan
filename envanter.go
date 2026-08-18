package main

import (
	"fmt"
	"strings"
)

func envanterCikar(ayar *Ayar, depo *Depo) error {
	cikti, hataCikti, err := sshCalistir(ayar.Zimbra, ayar.Zimbra.ProvKomut+" -l gad")
	if err != nil {
		return fmt.Errorf("gad: %s", ilkDolu(hataCikti, err.Error()))
	}
	depo.AlanAdlari = doluSatirlar(cikti)

	cikti2, hataCikti2, err := sshCalistir(ayar.Zimbra, ayar.Zimbra.ProvKomut+" -l gaa")
	if err != nil {
		return fmt.Errorf("gaa: %s", ilkDolu(hataCikti2, err.Error()))
	}
	epostalar := doluSatirlar(cikti2)

	var yeniler []string
	atlanan := 0
	for _, eposta := range epostalar {
		if ayar.hesapAtlanir(eposta) {
			atlanan++
			continue
		}
		if _, varMi := depo.Hesaplar[eposta]; varMi {
			continue
		}
		yeniler = append(yeniler, eposta)
	}

	metalar := topluMetaGetir(ayar, yeniler)
	for _, eposta := range yeniler {
		h := &Hesap{Eposta: eposta, AlanAdi: alanAdiBul(eposta), Durum: "bekliyor"}
		m, varMi := metalar[eposta]
		if !varMi {
			m = map[string]string{}
			if err := metaGetir(ayar, h); err != nil {
				logla("uyarı: %s meta alınamadı: %v", eposta, err)
			}
		}
		if len(m) > 0 {
			h.GorunenAd = m["displayName"]
			h.Ad = m["givenName"]
			h.Soyad = m["sn"]
			h.Kota = m["zimbraMailQuota"]
			h.Cos = m["zimbraCOSId"]
			h.SifreOzeti = m["userPassword"]
		}
		depo.mu.Lock()
		depo.Hesaplar[eposta] = h
		depo.mu.Unlock()
	}
	depo.kaydetKilitli()
	logla("envanter: %d alan adı, %d hesap bulundu, %d yeni eklendi, %d atlandı",
		len(depo.AlanAdlari), len(epostalar), len(yeniler), atlanan)
	return nil
}

func topluMetaGetir(ayar *Ayar, epostalar []string) map[string]map[string]string {
	sonuc := map[string]map[string]string{}
	if len(epostalar) == 0 || gProva {
		return sonuc
	}
	const parcaBoyu = 500
	for bas := 0; bas < len(epostalar); bas += parcaBoyu {
		son := bas + parcaBoyu
		if son > len(epostalar) {
			son = len(epostalar)
		}
		var b strings.Builder
		fmt.Fprintf(&b, "%s -l <<'KERVAN_SON'\n", ayar.Zimbra.ProvKomut)
		for _, e := range epostalar[bas:son] {
			fmt.Fprintf(&b, "ga %s displayName givenName sn zimbraMailQuota zimbraCOSId userPassword\n", e)
		}
		fmt.Fprintf(&b, "KERVAN_SON\n")
		cikti, _, err := sshCalistir(ayar.Zimbra, b.String())
		if err != nil {
			logla("uyarı: toplu envanter sorgusu başarısız, hesap hesap denenecek: %v", err)
			return sonuc
		}
		var aktif map[string]string
		for _, satir := range strings.Split(cikti, "\n") {
			if strings.HasPrefix(satir, "# name ") {
				eposta := strings.TrimSpace(strings.TrimPrefix(satir, "# name "))
				aktif = map[string]string{}
				sonuc[eposta] = aktif
				continue
			}
			if aktif == nil || strings.HasPrefix(satir, " ") || strings.HasPrefix(satir, "\t") {
				continue
			}
			i := strings.Index(satir, ":")
			if i <= 0 {
				continue
			}
			ad := strings.TrimSpace(satir[:i])
			deger := strings.TrimSpace(satir[i+1:])
			if _, varMi := aktif[ad]; !varMi {
				aktif[ad] = deger
			}
		}
	}
	return sonuc
}

func metaGetir(ayar *Ayar, h *Hesap) error {
	p := ayar.Zimbra.ProvKomut
	cikti, _, err := sshCalistir(ayar.Zimbra,
		fmt.Sprintf("%s ga %s displayName givenName sn zimbraMailQuota zimbraCOSId", p, tirnakla(h.Eposta)))
	if err != nil {
		return err
	}
	m := nitelikAyristir(cikti)
	h.GorunenAd = m["displayName"]
	h.Ad = m["givenName"]
	h.Soyad = m["sn"]
	h.Kota = m["zimbraMailQuota"]
	h.Cos = m["zimbraCOSId"]

	ciktiSifre, _, hataSifre := sshCalistir(ayar.Zimbra, fmt.Sprintf("%s -l ga %s userPassword", p, tirnakla(h.Eposta)))
	if hataSifre == nil {
		h.SifreOzeti = nitelikAyristir(ciktiSifre)["userPassword"]
	}
	return nil
}
