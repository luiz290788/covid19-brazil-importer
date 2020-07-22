package covid19brazilimporter

import "fmt"

type Page struct {
	ParentName string `json:"parentName"`
	ParentID   int    `json:"parentId"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
}

type DataListing struct {
	Name       string  `json:"name"`
	Cases      int     `json:"cases"`
	Deaths     int     `json:"deaths"`
	ParentName *string `json:"parentName,omitempty"`
}

func newDataListing(region *Region, parent *Region) *DataListing {
	var parentName *string
	if parent != nil {
		parentName = &parent.Name
	}
	return &DataListing{
		Name:       region.Name,
		Cases:      region.LastData.Cases,
		Deaths:     region.LastData.Deaths,
		ParentName: parentName,
	}
}

func buildIndexes(regions Regions) ([]*Page, map[int][]*DataListing, error) {
	pages := make([]*Page, len(regions)-1)
	dataListing := map[int][]*DataListing{}
	regionIndex := 0
	for _, region := range regions {
		// Ignore parent -1 because that is the root node
		if region.ParentID == -1 {
			continue
		}

		parent, parentFound := regions[region.ParentID]
		if !parentFound {
			return nil, nil, fmt.Errorf("Parent not found: ParentId %d", region.ParentID)
		}

		pages[regionIndex] = &Page{
			ID:         region.ID,
			Name:       region.Name,
			ParentID:   region.ParentID,
			ParentName: parent.Name,
		}

		parentListing, listingFound := dataListing[region.ParentID]

		if listingFound {
			dataListing[region.ParentID] = append(parentListing, newDataListing(region, parent))
		} else {
			dataListing[region.ParentID] = []*DataListing{newDataListing(region, parent)}
		}
		regionIndex++
	}
	return pages, dataListing, nil
}
