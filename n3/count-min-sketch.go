// count-min-sketch.go

package n3

import (
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/shenwei356/countminsketch"
)

//
// n3 count min sketch
// threadsafe wrapper allows use in multiple go-routines
// only exposes Update and Estimate methods
// from the underlying cms.
//
type N3CMS struct {
	cms *countminsketch.CountMinSketch
	sync.Mutex
	fileName string
}

//
// creates a new cms instance, will re-create from
// a previously saved instance if the given binary file version
// of the cms exists - if not will be created.
//
func NewN3CMS(fileName string) (*N3CMS, error) {

	if fileName == "" {
		return nil, errors.New("must supply a filename for the cms")
	}

	cms := &countminsketch.CountMinSketch{}
	epsilon := 0.0001
	delta := 0.9999

	cms, err := countminsketch.NewFromFile(fileName)
	if err != nil {
		log.Println("could not create cms from existing file, will create new.")
	} else {
		return &N3CMS{cms: cms, fileName: fileName}, nil
	}

	cms, err = countminsketch.NewWithEstimates(epsilon, delta)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create cms:")
	}

	n3cms := &N3CMS{cms: cms, fileName: fileName}

	return n3cms, nil

}

//
// return the count of how many times this has been seen before
// within the epsilon-delta error range
//
func (n3cms *N3CMS) Estimate(key string) uint64 {
	n3cms.Lock()
	defer n3cms.Unlock()
	return n3cms.cms.EstimateString(key)
}

//
// update the count for a given item
//
func (n3cms *N3CMS) Update(key string, count uint64) {
	n3cms.Lock()
	n3cms.cms.UpdateString(key, count)
	n3cms.Unlock()
}

//
// closing the cms attempts to save the
// cms to the given file
//
func (n3cms *N3CMS) Close() {
	_, err := n3cms.cms.WriteToFile(n3cms.fileName)
	if err != nil {
		log.Println("error writing cms to file: ", err)
	}
}
