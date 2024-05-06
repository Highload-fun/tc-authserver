package geo

import (
	"encoding/csv"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type Geo struct {
	cityById map[uint32]*City
	cityByIp *Tree
}

type City struct {
	Id      uint32
	Country string
	SubDiv1 string
	SubDiv2 string
	Name    string
}

func New(path string) *Geo {
	cityById := loadCities(filepath.Join(path, "GeoLite2-City-Locations-en.csv"))
	tree := loadIp4(filepath.Join(path, "GeoLite2-City-Blocks-IPv4.csv"), cityById)

	return &Geo{
		cityById: cityById,
		cityByIp: tree,
	}
}

func (g *Geo) GetCityByIp(ip net.IP) *City {
	return g.cityByIp.Find(ip)
}

func loadCities(filename string) map[uint32]*City {
	start := time.Now()
	defer func() {
		log.Printf("Cities loaded in %s", time.Now().Sub(start))
	}()

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvR := csv.NewReader(f)
	if _, err := csvR.Read(); err != nil {
		log.Fatal(err)
	}

	byId := map[uint32]*City{}
	for {
		row, err := csvR.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}

		id, err := strconv.ParseUint(row[0], 10, 32)
		if err != nil {
			log.Fatal(err)
		}

		city := &City{
			Id:      uint32(id),
			Country: row[5],
			SubDiv1: row[7],
			SubDiv2: row[9],
			Name:    row[10],
		}

		byId[city.Id] = city
	}

	return byId
}

func loadIp4(filename string, cityById map[uint32]*City) *Tree {
	start := time.Now()
	defer func() {
		log.Printf("Ip4 loaded in %s", time.Now().Sub(start))
	}()

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	csvR := csv.NewReader(f)
	if _, err := csvR.Read(); err != nil {
		log.Fatal(err)
	}

	tree := NewTree()

	for {
		row, err := csvR.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			log.Fatal(err)
		}

		if row[1] == "" {
			continue
		}

		geoId, err := strconv.ParseUint(row[1], 10, 32)
		if err != nil {
			log.Printf("%s: %s (%v)", row[0], row[1], err)
			continue
		}

		_, ipNet, err := net.ParseCIDR(row[0])
		if err != nil {
			log.Fatal(err)
		}

		city, exists := cityById[uint32(geoId)]
		if !exists {
			log.Printf("Cannot find geo id %d", geoId)
			continue
		}

		tree.Insert(ipNet, city)
	}

	return tree
}
