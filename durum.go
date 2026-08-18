package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

type Hesap struct {
	Eposta        string `json:"eposta"`
	AlanAdi       string `json:"alanAdi"`
	GorunenAd     string `json:"gorunenAd"`
	Ad            string `json:"ad"`
	Soyad         string `json:"soyad"`
	Cos           string `json:"cos"`
	Kota          string `json:"kota"`
	SifreOzeti    string `json:"sifreOzeti"`
	KaynakBayt    int64  `json:"kaynakBayt"`
	HedefBayt     int64  `json:"hedefBayt"`
	Durum         string `json:"durum"`
	Basarisiz     bool   `json:"basarisiz"`
	BasarisizAdim string `json:"basarisizAdim"`
	Hata          string `json:"hata"`
	SonEsitleme   string `json:"sonEsitleme"`
	Guncelleme    string `json:"guncelleme"`
}

type Depo struct {
	mu         sync.Mutex
	yol        string
	AlanAdlari []string          `json:"alanAdlari"`
	Hesaplar   map[string]*Hesap `json:"hesaplar"`
}

func yeniDepo(yol string) *Depo {
	return &Depo{yol: yol, Hesaplar: map[string]*Hesap{}}
}

func depoYukle(yol string) (*Depo, error) {
	d := yeniDepo(yol)
	b, err := os.ReadFile(yol)
	if err != nil {
		if os.IsNotExist(err) {
			return d, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(b, d); err != nil {
		return nil, err
	}
	if d.Hesaplar == nil {
		d.Hesaplar = map[string]*Hesap{}
	}
	d.yol = yol
	return d, nil
}

func (d *Depo) kaydet() error {
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	gecici := d.yol + ".tmp"
	if err := os.WriteFile(gecici, b, 0o600); err != nil {
		return err
	}
	return os.Rename(gecici, d.yol)
}

func (d *Depo) kaydetKilitli() {
	d.mu.Lock()
	defer d.mu.Unlock()
	_ = d.kaydet()
}

func (d *Depo) guncelle(eposta string, fn func(*Hesap)) {
	d.mu.Lock()
	defer d.mu.Unlock()
	h := d.Hesaplar[eposta]
	if h == nil {
		return
	}
	fn(h)
	h.Guncelleme = simdi()
	_ = d.kaydet()
}

func (d *Depo) listele() []*Hesap {
	d.mu.Lock()
	defer d.mu.Unlock()
	liste := make([]*Hesap, 0, len(d.Hesaplar))
	for _, h := range d.Hesaplar {
		liste = append(liste, h)
	}
	sort.Slice(liste, func(i, j int) bool { return liste[i].Eposta < liste[j].Eposta })
	return liste
}

func simdi() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
