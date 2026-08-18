package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type SunucuAyar struct {
	Adres          string   `json:"adres"`
	Port           int      `json:"port"`
	Kullanici      string   `json:"kullanici"`
	SSHAnahtar     string   `json:"sshAnahtar"`
	SSHEkSecenek   []string `json:"sshEkSecenekler"`
	PostaKullanici string   `json:"postaKullanici"`
	ProvKomut      string   `json:"provKomut"`
	PostaKomut     string   `json:"postaKomut"`
}

type DogrudanAyar struct {
	ZimbraAdres     string `json:"zimbraAdres"`
	ZimbraKullanici string `json:"zimbraKullanici"`
	ZimbraPort      int    `json:"zimbraPort"`
	CarbonioAnahtar string `json:"carbonioUzerindekiAnahtar"`
}

type EsitlemeAyar struct {
	CalistigiYer          string   `json:"calistigiYer"`
	ZimbraImapAdres       string   `json:"zimbraImapAdres"`
	ZimbraImapPort        int      `json:"zimbraImapPort"`
	ZimbraYonetici        string   `json:"zimbraYonetici"`
	ZimbraYoneticiSifre   string   `json:"zimbraYoneticiSifre"`
	CarbonioImapAdres     string   `json:"carbonioImapAdres"`
	CarbonioImapPort      int      `json:"carbonioImapPort"`
	CarbonioYonetici      string   `json:"carbonioYonetici"`
	CarbonioYoneticiSifre string   `json:"carbonioYoneticiSifre"`
	EkSecenekler          []string `json:"ekSecenekler"`
}

type Ayar struct {
	Zimbra              SunucuAyar        `json:"zimbra"`
	Carbonio            SunucuAyar        `json:"carbonio"`
	UzakCalismaDizini   string            `json:"uzakCalismaDizini"`
	EszamanliSayi       int               `json:"eszamanliSayi"`
	AktarimModu         string            `json:"aktarimModu"`
	AktarmaliYerelDizin string            `json:"aktarmaliYerelDizin"`
	Dogrudan            DogrudanAyar      `json:"dogrudan"`
	Esitleme            EsitlemeAyar      `json:"esitleme"`
	VarsayilanSifre     string            `json:"varsayilanSifre"`
	CosEsleme           map[string]string `json:"cosEsleme"`
	KotaUygula          bool              `json:"kotaUygula"`
	Temizlik            bool              `json:"temizlik"`
	AtlanacakHesaplar   []string          `json:"atlanacakHesaplar"`
	AtlanacakDesenler   []string          `json:"atlanacakDesenler"`
	DahilAlanlar        []string          `json:"dahilAlanlar"`
	DurumDosyasi        string            `json:"durumDosyasi"`
}

func ayarYukle(yol string) (*Ayar, error) {
	b, err := os.ReadFile(yol)
	if err != nil {
		return nil, err
	}
	var a Ayar
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, fmt.Errorf("ayar JSON: %w", err)
	}
	if a.Zimbra.Port == 0 {
		a.Zimbra.Port = 22
	}
	if a.Carbonio.Port == 0 {
		a.Carbonio.Port = 22
	}
	if a.Zimbra.ProvKomut == "" {
		a.Zimbra.ProvKomut = "zmprov"
	}
	if a.Carbonio.ProvKomut == "" {
		a.Carbonio.ProvKomut = "carbonio prov"
	}
	if a.Zimbra.PostaKomut == "" {
		a.Zimbra.PostaKomut = "zmmailbox"
	}
	if a.Carbonio.PostaKomut == "" {
		a.Carbonio.PostaKomut = "zmmailbox"
	}
	if a.UzakCalismaDizini == "" {
		a.UzakCalismaDizini = "/tmp/kervan"
	}
	if a.EszamanliSayi <= 0 {
		a.EszamanliSayi = 3
	}
	if a.AktarimModu == "" {
		a.AktarimModu = "aktarmali"
	}
	if a.AktarmaliYerelDizin == "" {
		a.AktarmaliYerelDizin = "."
	}
	if a.DurumDosyasi == "" {
		a.DurumDosyasi = "durum.json"
	}
	return &a, nil
}

func (a *Ayar) hesapAtlanir(eposta string) bool {
	kucuk := strings.ToLower(eposta)
	for _, h := range a.AtlanacakHesaplar {
		if strings.EqualFold(h, eposta) {
			return true
		}
	}
	for _, d := range a.AtlanacakDesenler {
		if d != "" && strings.Contains(kucuk, strings.ToLower(d)) {
			return true
		}
	}
	sistem := []string{"galsync", "ham.", "spam.", "virus-quarantine.", "admin@", "wiki@", "zmnginx"}
	for _, d := range sistem {
		if strings.Contains(kucuk, d) {
			return true
		}
	}
	if len(a.DahilAlanlar) > 0 {
		alan := alanAdiBul(eposta)
		bulundu := false
		for _, d := range a.DahilAlanlar {
			if strings.EqualFold(d, alan) {
				bulundu = true
				break
			}
		}
		if !bulundu {
			return true
		}
	}
	return false
}
