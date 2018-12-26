package distmatrix

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
)

type DistMatrix struct {
	values [][]int
}

func NewDM(count int) *DistMatrix {
	var dm DistMatrix
	dm.values = make([][]int, count)
	return &dm
}

func (d *DistMatrix) addA(arr []int, ix int) {
	if len(arr) < 1 {
		return
	}
	d.values[ix] = arr
}

func (d *DistMatrix) dist(a, b int) (int, error) {
	if a == b {
		return 0, nil
	}

	maxa := len(d.values)
	if a > maxa || b > maxa {
		return 0, errors.New("A or B is greater than max accounts")
	}
	lower := a
	higher := b
	if a > b {
		lower = b
		higher = a
	}
	return d.values[lower][higher-lower], nil
}

func (d *DistMatrix) belowDistance(pos []int, cutoff int) (int, error) {
	less := 0
	for ix := 0; ix < len(pos); ix++ {
		subpos := pos[ix:]
		for _, v := range subpos {
			dist, err := d.dist(pos[ix], v)
			if err != nil {
				return 0, err
			}
			if dist == 0 {
				continue
			}
			//log.Infof("Between %d and %d, dist is %d", pos[ix], v, dist)
			if dist < cutoff {
				less++
			}
		}
	}
	return less, nil
}

func (d *DistMatrix) multiMinMax(pos []int) (int, int, error) {
	max := 0
	min := -1

	for ix := 0; ix < len(pos); ix++ {
		subpos := pos[ix:]
		for _, v := range subpos {
			dist, err := d.dist(pos[ix], v)
			if err != nil {
				return 0, 0, err
			}
			if dist == 0 {
				continue
			}
			if dist > max {
				max = dist
			}
			if min == -1 {
				min = dist
			}
			if dist < min {
				min = dist
			}
		}
	}
	return min, max, nil
}

func (d *DistMatrix) print() {
	for ix, arr := range d.values {
		log.Debugf("%d: %+v", ix, arr)
	}
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Infof("%s took %s", name, elapsed)
}
