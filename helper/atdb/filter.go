package atdb

import (
	"strconv"
	"time"

	"github.com/gocroot/helper/atapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetDateSekarang() (datesekarang time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")

	t := time.Now().In(location) //.Truncate(24 * time.Hour)
	datesekarang = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func TodayFilter() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateSekarang()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateSekarang().Add(24 * time.Hour)),
	}
}

func YesterdayNotLiburFilter() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateKemarinBukanHariLibur()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateKemarinBukanHariLibur().Add(24 * time.Hour)),
	}
}

func YesterdayFilter() bson.M {
	return bson.M{
		"$gte": primitive.NewObjectIDFromTimestamp(GetDateKemarin()),
		"$lt":  primitive.NewObjectIDFromTimestamp(GetDateKemarin().Add(24 * time.Hour)),
	}
}

func GetDateKemarinBukanHariLibur() (datekemarinbukanlibur time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")
	n := -1
	t := time.Now().AddDate(0, 0, n).In(location) //.Truncate(24 * time.Hour)
	for HariLibur(t) {
		n -= 1
		t = time.Now().AddDate(0, 0, n).In(location)
	}

	datekemarinbukanlibur = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func GetDateKemarin() (datekemarin time.Time) {
	// Definisi lokasi waktu sekarang
	location, _ := time.LoadLocation("Asia/Jakarta")
	n := -1
	t := time.Now().AddDate(0, 0, n).In(location) //.Truncate(24 * time.Hour)
	datekemarin = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())

	return
}

func HariLibur(thedate time.Time) (libur bool) {
	wekkday := thedate.Weekday()
	inhari := int(wekkday)
	if inhari == 0 || inhari == 6 {
		libur = true
	}
	tglskr := thedate.Format("2006-01-02")
	tgl := int(thedate.Month())
	urltarget := "https://dayoffapi.vercel.app/api?month=" + strconv.Itoa(tgl)
	_, hasil, _ := atapi.Get[[]NewLiburNasional](urltarget)
	for _, v := range hasil {
		if v.Tanggal == tglskr {
			libur = true
		}
	}
	return
}
