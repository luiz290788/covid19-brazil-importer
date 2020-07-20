package covid19brazilimporter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const stateCol = 1
const cityCol = 2
const codUFCol = 3
const codMunCol = 4
const healthRegionIDCol = 5
const healthRegionNameCol = 6
const populationCol = 9

type RegionID = int

type Region struct {
	ID               RegionID     `json:"id"`
	Name             string       `json:"name"`
	ParentID         RegionID     `json:"parentId,omitempty"`
	Group            *string      `json:"group,omitempty"`
	HealthRegionID   int          `json:"healthRegionId,omitempty"`
	HealthRegionName string       `json:"healthRegionName,omitempty"`
	Population       int          `json:"population,omitempty"`
	LastData         *DataEntry   `json:"lastData,omitempty"`
	Timeseries       []*DataEntry `json:"timeseries,omitempty"`
}

const dateCol = 7
const weekCol = 8
const casesCol = 10
const newCasesCol = 11
const deathsCol = 12
const newDeathsCol = 13
const recoveredCol = 14
const underTreatmentCol = 15

type DataEntry struct {
	Date           time.Time `json:"date"`
	Week           int       `json:"week"`
	Cases          int       `json:"cases"`
	NewCases       int       `json:"newCases"`
	Deaths         int       `json:"deaths"`
	NewDeaths      int       `json:"newDeaths"`
	Recovered      int       `json:"recovered"`
	UnderTreatment int       `json:"underTreatment"`
}

type Regions map[RegionID]*Region

func getOrDefault(row []*string, index int, defaultValue *string) *string {
	if len(row) > index && index >= 0 {
		return row[index]
	}
	return defaultValue
}

func readDataEntry(row []*string) (*DataEntry, error) {
	// var err error = nil
	var date time.Time
	var week, cases, newCases, deaths, newDeaths, recovered, underTreatment int

	defaultValue := "0"

	date, _ = parseDate(*getOrDefault(row, dateCol, &defaultValue))
	week, _ = strconv.Atoi(*getOrDefault(row, weekCol, &defaultValue))
	cases, _ = strconv.Atoi(*getOrDefault(row, casesCol, &defaultValue))
	newCases, _ = strconv.Atoi(*getOrDefault(row, newCasesCol, &defaultValue))
	deaths, _ = strconv.Atoi(*getOrDefault(row, deathsCol, &defaultValue))
	newDeaths, _ = strconv.Atoi(*getOrDefault(row, newDeathsCol, &defaultValue))
	// Use 0 if error
	recovered, _ = strconv.Atoi(*getOrDefault(row, recoveredCol, &defaultValue))
	// Use 0 if error
	underTreatment, _ = strconv.Atoi(*getOrDefault(row, underTreatmentCol, &defaultValue))

	return &DataEntry{
		Date:           date,
		Week:           week,
		Cases:          cases,
		NewCases:       newCases,
		Deaths:         deaths,
		NewDeaths:      newDeaths,
		Recovered:      recovered,
		UnderTreatment: underTreatment,
	}, nil
}

func parsePopulation(strPopulation *string) (int, error) {
	expression := regexp.MustCompile(`^([0-9.]+)\s*(\(\d+\))?$`)
	result := expression.FindStringSubmatch(*strPopulation)
	return strconv.Atoi(strings.ReplaceAll(result[1], ".", ""))
}

func (region *Region) addData(data *DataEntry) error {
	if region.LastData == nil || region.LastData.Date.Before(data.Date) {
		region.LastData = data
	}
	region.Timeseries = append(region.Timeseries, data)
	return nil
}

func (regions Regions) addCountry(row []*string) error {
	var codUF RegionID
	var err error
	codUF, err = strconv.Atoi(*row[codUFCol])
	if err != nil {
		return err
	}

	var region *Region
	data, err := readDataEntry(row)
	if err != nil {
		return err
	}
	var ok bool
	if region, ok = regions[codUF]; ok {
		region.addData(data)
		return nil
	}

	var population int
	population, err = parsePopulation(row[populationCol])
	if err != nil {
		return err
	}

	region = &Region{
		ID:         codUF,
		Name:       *row[0],
		ParentID:   -1,
		Population: population,
		Timeseries: []*DataEntry{},
	}
	regions[codUF] = region
	region.addData(data)
	return nil
}

func (regions Regions) addCity(row []*string) error {
	var codMun RegionID
	var err error
	codMun, err = strconv.Atoi(*row[codMunCol])
	if err != nil {
		return fmt.Errorf("error while parsing codMun: %s", err.Error())
	}

	var region *Region
	data, err := readDataEntry(row)
	if err != nil {
		return err
	}

	var ok bool
	if region, ok = regions[codMun]; ok {
		region.addData(data)
		return nil
	}

	var codUF, healthRegionID, population int
	codUF, err = strconv.Atoi(*row[codUFCol])
	if err != nil {
		return fmt.Errorf("error while parsing codUF: %s", err.Error())
	}
	population, err = parsePopulation(row[populationCol])
	if err != nil {
		return fmt.Errorf("error while parsing population: %s", err.Error())
	}

	region = &Region{
		ID:               codMun,
		Name:             *row[cityCol],
		ParentID:         codUF,
		Group:            row[0],
		HealthRegionID:   healthRegionID,
		HealthRegionName: *row[healthRegionNameCol],
		Population:       population,
		Timeseries:       []*DataEntry{},
	}
	region.addData(data)
	regions[codMun] = region
	return nil
}

func (regions Regions) addOther(row []*string) error {
	var codMun RegionID
	var err error
	codMun, err = strconv.Atoi(*row[codMunCol])
	if err != nil {
		return fmt.Errorf("error while parsing codMun: %s", err.Error())
	}

	var region *Region
	data, err := readDataEntry(row)
	if err != nil {
		return err
	}

	var ok bool
	if region, ok = regions[codMun]; ok {
		region.addData(data)
		return nil
	}

	var codUF int
	codUF, err = strconv.Atoi(*row[codUFCol])
	if err != nil {
		return fmt.Errorf("error parsing coduf: %s", err.Error())
	}

	region = &Region{
		ID:         codMun,
		Name:       "Other",
		ParentID:   codUF,
		Group:      row[0],
		Timeseries: []*DataEntry{},
	}
	region.addData(data)
	regions[codMun] = region
	return nil
}

func (regions Regions) addState(row []*string) error {
	var codUF RegionID
	var err error
	codUF, err = strconv.Atoi(*row[codUFCol])
	if err != nil {
		return err
	}

	var region *Region
	data, err := readDataEntry(row)
	if err != nil {
		return err
	}

	var ok bool
	if region, ok = regions[codUF]; ok {
		region.addData(data)
		return nil
	}

	var population int
	population, err = strconv.Atoi(*row[populationCol])
	if err != nil {
		return err
	}

	region = &Region{
		ID:         codUF,
		Name:       *row[stateCol],
		ParentID:   76, // fixed to Brazil
		Group:      row[0],
		Population: population,
		Timeseries: []*DataEntry{},
	}
	region.addData(data)
	regions[codUF] = region
	return nil
}

func (regions Regions) ProcessRow(row []*string) error {
	state := row[stateCol]
	if state == nil || len(*state) == 0 {
		err := regions.addCountry(row)
		if err != nil {
			return fmt.Errorf("error while adding country: %s", err.Error())
		}
		return nil
	}

	city := row[cityCol]
	if city != nil && len(*city) != 0 {
		err := regions.addCity(row)
		if err != nil {
			return fmt.Errorf("error while adding city: %s", err.Error())
		}
		return nil
	}

	codMun := row[codMunCol]
	if codMun != nil && len(*codMun) != 0 {
		err := regions.addOther(row)
		if err != nil {
			return fmt.Errorf("error while adding other: %s", err.Error())
		}
		return nil
	}

	err := regions.addState(row)
	if err != nil {
		return fmt.Errorf("error while adding state: %s", err.Error())
	}
	return nil
}
