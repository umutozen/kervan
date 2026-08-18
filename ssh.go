package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

var gProva bool

func sshTemelArg(s SunucuAyar) []string {
	arg := []string{
		"-p", strconv.Itoa(s.Port),
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
	}
	if s.SSHAnahtar != "" {
		arg = append(arg, "-i", s.SSHAnahtar)
	}
	arg = append(arg, s.SSHEkSecenek...)
	return arg
}

func scpTemelArg(s SunucuAyar) []string {
	arg := []string{
		"-P", strconv.Itoa(s.Port),
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=accept-new",
	}
	if s.SSHAnahtar != "" {
		arg = append(arg, "-i", s.SSHAnahtar)
	}
	arg = append(arg, s.SSHEkSecenek...)
	return arg
}

func uzakSarmala(s SunucuAyar, betik string) string {
	b64 := base64.StdEncoding.EncodeToString([]byte(betik))
	kabuk := "bash -l"
	if s.PostaKullanici != "" {
		kabuk = "su - " + s.PostaKullanici + " -s /bin/bash"
	}
	return fmt.Sprintf("printf %%s %s | base64 -d | %s", b64, kabuk)
}

func sshCalistir(s SunucuAyar, betik string) (string, string, error) {
	if gProva {
		logla("[PROVA] SSH %s@%s:\n%s", s.Kullanici, s.Adres, girintile(betik))
		return "", "", nil
	}
	arg := sshTemelArg(s)
	arg = append(arg, fmt.Sprintf("%s@%s", s.Kullanici, s.Adres), uzakSarmala(s, betik))
	komut := exec.Command("ssh", arg...)
	var cikti, hataCikti bytes.Buffer
	komut.Stdout = &cikti
	komut.Stderr = &hataCikti
	err := komut.Run()
	return strings.TrimRight(cikti.String(), "\r\n"), strings.TrimRight(hataCikti.String(), "\r\n"), err
}

func scpIndir(s SunucuAyar, uzak, yerel string) error {
	if gProva {
		logla("[PROVA] scp %s@%s:%s -> %s", s.Kullanici, s.Adres, uzak, yerel)
		return nil
	}
	arg := scpTemelArg(s)
	arg = append(arg, fmt.Sprintf("%s@%s:%s", s.Kullanici, s.Adres, uzak), yerel)
	return yerelCalistir("scp", arg...)
}

func scpYukle(s SunucuAyar, yerel, uzak string) error {
	if gProva {
		logla("[PROVA] scp %s -> %s@%s:%s", yerel, s.Kullanici, s.Adres, uzak)
		return nil
	}
	arg := scpTemelArg(s)
	arg = append(arg, yerel, fmt.Sprintf("%s@%s:%s", s.Kullanici, s.Adres, uzak))
	return yerelCalistir("scp", arg...)
}

func yerelCalistir(ad string, arg ...string) error {
	komut := exec.Command(ad, arg...)
	var hataCikti bytes.Buffer
	komut.Stderr = &hataCikti
	if err := komut.Run(); err != nil {
		return fmt.Errorf("%s: %v: %s", ad, err, strings.TrimSpace(hataCikti.String()))
	}
	return nil
}

func girintile(s string) string {
	satirlar := strings.Split(s, "\n")
	for i := range satirlar {
		satirlar[i] = "    " + satirlar[i]
	}
	return strings.Join(satirlar, "\n")
}
