package distmatrix

import (
	"errors"
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

type Collection struct {
	id     string // unique collection id for use in a cache, if needed
	dm     *DistMatrix
	points map[string]Point
}

type Point struct {
	id       string
	position int
	lat      float64
	long     float64
}

func NewCollection(id string) *Collection {
	var na Collection
	na.id = id
	na.points = make(map[string]Point)
	return &na
}

func (c *Collection) Add(id string, lat, long float64) {
	point := Point{id: id, lat: lat, long: long}
	point.position = len(c.points)
	c.points[id] = point
	c.dm = nil
}

func (c *Collection) Dist(a, b string) (int, error) {
	var err error
	dist := 0
	c.buildDistMatrix()

	if a != "" && b != "" {
		ap, ea := c.retPos(a)
		bp, eb := c.retPos(b)
		if ea != nil || eb != nil {
			log.Errorf("Errors: id not in collection")
			return 0, errors.New("Collection:Dist: id not found")
		} else {
			dist, err = c.dm.dist(ap, bp)
			if err != nil {
				log.Errorf("dist func ret error: %s", err.Error())
			} else {
				log.Infof("A to B distance is %d", dist)
			}
		}
	}
	return dist, nil
}

func (c *Collection) MinMax() (int, int, error) {
	c.buildDistMatrix()
	clen := len(c.points)
	plist := make([]int, clen)
	for ix := 0; ix < clen; ix++ {
		plist[ix] = ix
	}
	return c.dm.multiMinMax(plist)
}

func (c *Collection) MinMaxSS(plist []string) (int, int, error) {
	c.buildDistMatrix()
	posa, e := c.retPosArray(plist)
	if e != nil {
		return 0, 0, e
	}
	return c.dm.multiMinMax(posa)
}

func (c *Collection) CountLessThan(gap int) (int, error) {
	c.buildDistMatrix()
	clen := len(c.points)
	plist := make([]int, clen)
	for ix := 0; ix < clen; ix++ {
		plist[ix] = ix
	}
	return c.dm.belowDistance(plist, gap)
}

func (c *Collection) CountSSLessThan(plist []string, gap int) (int, error) {
	c.buildDistMatrix()
	posa, e := c.retPosArray(plist)
	if e != nil {
		return 0, e
	}
	return c.dm.belowDistance(posa, gap)
}

func (c *Collection) buildDistMatrix() {
	if c.dm != nil {
		return
	}

	defer timeTrack(time.Now(), "retDistanceMetrix")

	acnt := len(c.points)
	dm := NewDM(acnt)

	// make a list
	allac := make([]Point, acnt)
	for _, oacc := range c.points {
		allac[oacc.position] = oacc
		log.Debugf("Position: %d, Point: %+v", oacc.position, oacc)
	}

	for aix := 0; aix < acnt; aix++ {
		alen := acnt - aix
		darr := make([]int, alen)

		curac := allac[aix]
		la1 := curac.lat
		lo1 := curac.long

		subl := allac[aix:]
		for k, v := range subl {
			if k == 0 {
				darr[k] = 0
				continue
			}
			la2 := v.lat
			lo2 := v.long
			darr[k] = int(Dist(la1, lo1, la2, lo2))
		}
		dm.addA(darr, aix)
	}
	c.dm = dm
}

func (c *Collection) retPos(hk string) (int, error) {
	if acc, k := c.points[hk]; !k {
		log.Errorf("id: %s not in points", hk)
		return 0, errors.New("not found")
	} else {
		return acc.position, nil
	}
}

func (c *Collection) retPosArray(hks []string) ([]int, error) {
	var retarr []int
	for _, hk := range hks {
		if acc, k := c.points[hk]; !k {
			log.Errorf("id: %s not in points", hk)
		} else {
			retarr = append(retarr, acc.position)
		}
	}
	return retarr, nil
}

func (c *Collection) print() {
	fmt.Println("*************************************************************")
	fmt.Printf("*** Node ID: %s", c.id)
	fmt.Println("*** *** Account(s) *** ***")
	for aid, acc := range c.points {
		fmt.Printf("***< %d > %s: Lat: %f, Long: %f", acc.position, aid, acc.lat, acc.long)
	}
}

func Dist(lat1, lon1, lat2, lon2 float64) float64 {
	//defer timeTrack(time.Now(), "distFunction")

	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}
