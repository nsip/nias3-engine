// count-min-sketch.go

package n3

import (
	"log"
	"sync"

	"github.com/pkg/errors"
	"github.com/shenwei356/countminsketch"
)

type N3CMS struct {
	cms *countminsketch.CountMinSketch
	sync.Mutex
	fileName string
}

func NewN3CMS(fileName string) (*N3CMS, error) {

	if fileName == "" {
		return nil, errors.New("must supply a filename for the cms")
	}

	cms := &countminsketch.CountMinSketch{}
	epsilon := 0.0001
	delta := 0.9999

	cms, err := countminsketch.NewFromFile(fileName)
	if err != nil {
		log.Println("could not open cms file, will create new.")
	} else {
		return &N3CMS{cms: cms, fileName: fileName}, nil
	}

	cms, err = countminsketch.NewWithEstimates(epsilon, delta)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create cms:")
	}

	n3cms := &N3CMS{cms: cms, fileName: fileName}

	log.Printf("\t New CMS:\n\n%+v\n\n", n3cms)

	return n3cms, nil

}

func (n3cms *N3CMS) Estimate(key string) uint64 {
	n3cms.Lock()
	defer n3cms.Unlock()
	return n3cms.cms.EstimateString(key)
}

func (n3cms *N3CMS) Update(key string, count uint64) {
	n3cms.Lock()
	n3cms.cms.UpdateString(key, count)
	n3cms.Unlock()
}

func (n3cms *N3CMS) Close() {
	_, err := n3cms.cms.WriteToFile(n3cms.fileName)
	if err != nil {
		log.Println("error writing cms to file: ", err)
	}
}
