package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

func rastgeleParola() string {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "K-" + simdi()
	}
	return "K-" + base64.RawURLEncoding.EncodeToString(b)
}

func logla(bicim string, a ...any) {
	log.Printf(bicim, a...)
}

func denetle(err error) {
	if err != nil {
		log.Fatalf("hata: %v", err)
	}
}

func tirnakla(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func nitelikCiftleri(kv ...string) string {
	var b strings.Builder
	for i := 0; i+1 < len(kv); i += 2 {
		ad, deger := kv[i], kv[i+1]
		if deger == "" {
			continue
		}
		fmt.Fprintf(&b, " %s %s", ad, tirnakla(deger))
	}
	return b.String()
}

func ilkDolu(degerler ...string) string {
	for _, d := range degerler {
		if strings.TrimSpace(d) != "" {
			return d
		}
	}
	return ""
}

func yerelKisim(eposta string) string {
	if i := strings.Index(eposta, "@"); i > 0 {
		return eposta[:i]
	}
	return eposta
}

func alanAdiBul(eposta string) string {
	if i := strings.Index(eposta, "@"); i >= 0 && i+1 < len(eposta) {
		return eposta[i+1:]
	}
	return ""
}

func sonSatir(s string) string {
	satirlar := strings.Split(strings.TrimRight(s, "\r\n"), "\n")
	return satirlar[len(satirlar)-1]
}

func doluSatirlar(s string) []string {
	var liste []string
	for _, satir := range strings.Split(s, "\n") {
		satir = strings.TrimSpace(satir)
		if satir != "" {
			liste = append(liste, satir)
		}
	}
	return liste
}

func icerir(liste []string, deger string) bool {
	for _, x := range liste {
		if strings.EqualFold(x, deger) {
			return true
		}
	}
	return false
}

func virgulAyir(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var liste []string
	for _, parca := range strings.Split(s, ",") {
		if parca = strings.TrimSpace(parca); parca != "" {
			liste = append(liste, parca)
		}
	}
	return liste
}

func temizAd(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", " ", "_")
	return r.Replace(s)
}

func uzakTgz(ayar *Ayar, eposta string) string {
	return strings.TrimRight(ayar.UzakCalismaDizini, "/") + "/" + temizAd(eposta) + ".tgz"
}

func nitelikAyristir(cikti string) map[string]string {
	m := map[string]string{}
	for _, satir := range strings.Split(cikti, "\n") {
		if strings.HasPrefix(satir, " ") || strings.HasPrefix(satir, "\t") {
			continue
		}
		i := strings.Index(satir, ":")
		if i <= 0 {
			continue
		}
		ad := strings.TrimSpace(satir[:i])
		deger := strings.TrimSpace(satir[i+1:])
		if _, varMi := m[ad]; !varMi {
			m[ad] = deger
		}
	}
	return m
}
