# Kervan

```
 _  __ _____  ____  __     __    _     _   _
| |/ /| ____||  _ \ \ \   / /   / \   | \ | |
| ' / |  _|  | |_) | \ \ / /   / _ \  |  \| |
| . \ | |___ |  _ <   \ V /   / ___ \ | |\  |
|_|\_\|_____||_| \_\   \_/   /_/   \_\|_| \_|

     Göçtü kervan kaldık dağlar başında..
```

Zimbra'dan Carbonio'ya geçiyorsanız bu araç işinizi görür. Hesapları,
şifreleri, postaları, kişileri ve takvimleri eski sunucudan yenisine taşır.

## Nasıl çalışıyor?

Kervan sizin yerinize sunuculara bağlanıp komut çalıştıran bir yardımcıdır.
Kendi bilgisayarınızda durur, iki sunucuya da SSH ile bağlanır ve şu sırayla
ilerler:

1. **Önce Zimbra'ya bakar.** Kaç hesap var, kimin adı ne, hepsini bir listeye
   çıkarır. Şifrelerin sunucuda tutulan şifrelenmiş halini de alır.
2. **Carbonio'da hesapları açar.** Şifreler aynen taşındığı için
   kullanıcılarınız yeni sunucuya eski şifreleriyle girer. Kimseye "şifreniz
   sıfırlandı" maili atmak zorunda kalmazsınız.
3. **Postaları taşır.** Her hesabın postası, kişileri, takvimi, görevleri ve
   klasör düzeni tek dosya halinde Zimbra'dan çıkar, Carbonio'ya girer.
4. **Nerede kaldığını unutmaz.** Her hesabın hangi adımda olduğu `durum.json`
   dosyasına işlenir. Bağlantı kopsa, bilgisayar kapansa bile aynı komutu
   tekrar çalıştırın; kaldığı yerden devam eder, hiçbir şeyi iki kez taşımaz.

## Kurulum

Üç şey lazım: Go, ssh ve bu depo.

**Go kurulu mu?** Terminale `go version` yazın. Cevap geliyorsa kurulu,
gelmiyorsa:

- Windows: `winget install GoLang.Go` (ya da https://go.dev/dl adresinden indirin)
- Ubuntu / Debian: `sudo apt install golang-go`
- Rocky / RHEL / AlmaLinux: `sudo dnf install golang`

**ssh zaten vardır.** Windows 10 ve üzeri, Linux ve Mac'te hazır gelir.

**Depoyu indirip derleyin:**

```bash
git clone https://github.com/umutozen/kervan.git
cd kervan
go build -o kervan .
```

Hepsi bu. Ortaya tek bir program dosyası çıkar; veritabanı, kütüphane, başka
hiçbir şey kurmanız gerekmez.

## Sunuculara erişim

Kervan sunuculara şifreyle değil SSH anahtarıyla bağlanır. Anahtarınız yoksa
üretip iki sunucuya da tanıtın:

```bash
ssh-keygen -t ed25519 -f ~/.ssh/kervan_anahtar -N ""
ssh-copy-id -i ~/.ssh/kervan_anahtar.pub zimbra@ZIMBRA_ADRESI
ssh-copy-id -i ~/.ssh/kervan_anahtar.pub zextras@CARBONIO_ADRESI
```

Bağlandığınız kullanıcı önemli: Zimbra'da `zimbra`, Carbonio'da `zextras`
kullanıcısını kullanın. Root ile bağlanmak zorundaysanız ayar dosyasındaki
`postaKullanici` alanını doldurun, Kervan komutları o kullanıcıya geçerek
çalıştırır.

## Ayarlar

Örnek dosyayı kopyalayın, kendi bilgilerinizi yazın:

```bash
cp ayar.ornek.json ayar.json
```

Dosyayı açınca ne olduğu bellidir: `zimbra` ve `carbonio` bölümlerine sunucu
adresi, kullanıcı adı ve anahtar dosyanızın yolu yazılır. Gerisi olduğu gibi
kalabilir. Yine de bilmekte fayda olanlar:

- `eszamanliSayi`: aynı anda kaç hesap taşınsın. 3 iyi bir başlangıç;
  sunucuları yormak istemiyorsanız düşürün.
- `aktarimModu`: postalar normalde önce sizin bilgisayara iner, sonra
  Carbonio'ya çıkar (`aktarmali`). İki sunucu birbirini görüyorsa bunu
  `dogrudan` yapın; dosyalar sunucudan sunucuya gitsin, iş epey hızlanır.
  Bunun için Carbonio'dan Zimbra'ya da anahtarla ssh gerekir.
- `varsayilanSifre`: boş bırakın. Hesap açılırken gereken geçici şifre her
  hesap için rastgele üretilir ve zaten hemen ardından gerçek şifreyle ezilir.
- `atlanacakHesaplar` ve `atlanacakDesenler`: taşınmasını istemediğiniz
  hesaplar. Sistem hesapları (galsync, spam gibi) zaten kendiliğinden atlanır.
- `temizlik`: `true` yaparsanız taşıma biten hesabın geçici dosyaları
  sunuculardan silinir.

## Kullanım

Sıra şöyle:

```bash
./kervan kontrol      # her şey yerinde mi diye bakar (önce bunu çalıştırın)
./kervan envanter     # Zimbra'daki hesapları listeye çıkarır
./kervan tasi         # taşımayı başlatır
./kervan durum        # kim ne durumda, tablo halinde gösterir
./kervan dogrula      # taşınan dosya boyutlarını karşılaştırır
```

`kontrol` komutu taşımaya başlamadan sizin yerinize etrafa bakınır: iki
sunucuya bağlanabiliyor muyum, gereken komutlar yerinde mi, çalışma dizinine
yazabiliyor muyum, disk yeterli mi. Bir sorun varsa taşımanın ortasında değil,
en başta öğrenirsiniz.

Bir hesap hata verirse diğerleri etkilenmez, taşıma devam eder. İş bitince
hata verenleri toplayın:

```bash
./kervan tasi -yeniden    # sadece hatalıları, kaldıkları adımdan tekrar dener
```

Başlamadan önce ne yapacağını görmek isterseniz:

```bash
./kervan tasi -prova      # hiçbir şey yapmaz, yapacaklarını ekrana yazar
```

Sadece birkaç hesap taşımak için:

```bash
./kervan tasi -sadece evliya.celebi@ornek.com,gezgin@ornek.com
```

Ekranda her hesabın ilerleyişini adım adım görürsünüz:

```
OK    evliya.celebi@ornek.com  -> hesap açıldı
OK    evliya.celebi@ornek.com  -> dışa aktarıldı
OK    evliya.celebi@ornek.com  -> taşındı
OK    evliya.celebi@ornek.com  -> içe aktarıldı
BİTTİ evliya.celebi@ornek.com
```

## Tavsiyemiz

1. Önce `envanter` çekin, oluşan `rapor.csv`'yi açıp listeye bir bakın.
2. Toplu taşımadan önce bir iki test hesabını `-sadece` ile taşıyın,
   Carbonio'da o hesapla girip her şey yerinde mi kontrol edin.
3. Sonra hepsini taşıyın. Bu sırada kullanıcılar Zimbra'da çalışmaya devam
   edebilir, kimsenin haberi bile olmaz.
4. Geçiş günü `./kervan esitle` ile son gelen postaları eşitleyin ve DNS'teki
   MX kaydını en son değiştirin.

## Geçiş günü: son postaları eşitleme

Taşıma saatler sürebilir ve bu sırada Zimbra'ya posta gelmeye devam eder.
`esitle` komutu bu farkı kapatır: taşınmış her hesap için imapsync çalıştırır
ve sadece eksik postaları getirir. İstediğiniz kadar tekrar çalıştırabilirsiniz,
olan postayı bir daha kopyalamaz.

```bash
./kervan esitle
./kervan esitle -sadece evliya.celebi@ornek.com
```

Bunun için iki şey gerekir:

1. Eşitlemenin çalışacağı sunucuda (varsayılan Carbonio) imapsync kurulu
   olmalı: `apt install imapsync` ya da `dnf install imapsync`
2. Ayar dosyasındaki `esitleme` bölümüne iki sunucunun IMAP yönetici hesabı ve
   şifresi yazılmalı. Yönetici hesabı üzerinden bağlanıldığı için kullanıcı
   şifreleri yine gerekmez.

## Neler taşınır, neler taşınmaz?

Taşınır: posta, kişiler, takvim, görevler, klasörler, etiketler, şifreler,
kota, ad soyad.

Sizin ilgilenmeniz gerekenler: alias'lar (ikinci adresler), dağıtım listeleri,
COS tanımları, takvim paylaşım izinleri ve Briefcase dosyaları.

Son bir not: `carbonio prov` ve `zmmailbox` komutları sürümden sürüme ufak
farklılık gösterebilir, o yüzden test hesabı tavsiyesini ciddiye alın. İçiniz
rahat olsun; Kervan'ın Zimbra üzerinde çalıştırdığı komutların hepsi salt
okunurdur, kaynak sunucunuza zarar vermez.
