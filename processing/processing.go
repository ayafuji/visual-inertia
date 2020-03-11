package processing

import (
	"math"
)

const (
	WINDOW_SIZE = 10

	XYZ_AX_SCALE = 20
)

type DataSeries struct {
	data          []float64
	index_counter int

	current float64
	pre     float64
}

func NewDataSeries() *DataSeries {
	ds := DataSeries{}
	return &ds
}

func (d *DataSeries) AddData(data float64) {

	// size: 10 WindowSize 10
	if len(d.data) >= WINDOW_SIZE {
		d.data[d.index_counter] = data
		d.index_counter++

		// index_counter: 10 WindowSize 9
		if d.index_counter > WINDOW_SIZE-1 {
			d.index_counter = 0
		}
	} else {
		d.data = append(d.data, data)
	}

	d.pre = d.current
	d.current = data

}

func (d *DataSeries) GetMovingAverage() (float64, error) {

	var sum float64
	for _, val := range d.data {
		sum += float64(val)
	}

	result := sum / float64(len(d.data))
	return result, nil
}

func (d *DataSeries) GetAX() (float64, error) {
	return math.Abs(d.current-d.pre) * XYZ_AX_SCALE, nil
}

func (d *DataSeries) GetLatestData() (float64, error) {
	return d.current, nil
}

type XYZ struct {
	X *DataSeries
	Y *DataSeries
	Z *DataSeries
}

func NewXYZ() *XYZ {
	xzy := XYZ{X: NewDataSeries(), Y: NewDataSeries(), Z: NewDataSeries()}
	return &xzy
}

func (d *XYZ) AddData(x, y, z float64) {
	d.X.AddData(x)
	d.Y.AddData(y)
	d.Z.AddData(z)
}

func (d *XYZ) GetXAX() float64 {
	if val, err := d.X.GetAX(); err != nil {
		return 0
	} else {
		return val
	}
}

func (d *XYZ) GetYAX() float64 {
	if val, err := d.Y.GetAX(); err != nil {
		return 0
	} else {
		return val
	}
}

func (d *XYZ) GetZAX() float64 {
	if val, err := d.Z.GetAX(); err != nil {
		return 0
	} else {
		return val
	}
}

func (d *XYZ) GetXMA() float64 {
	if val, err := d.X.GetMovingAverage(); err != nil {
		return 0
	} else {
		return val
	}
}

func (d *XYZ) GetYMA() float64 {
	if val, err := d.Y.GetMovingAverage(); err != nil {
		return 0
	} else {
		return val
	}
}

func (d *XYZ) GetZMA() float64 {
	if val, err := d.Z.GetMovingAverage(); err != nil {
		return 0
	} else {
		return val
	}
}
