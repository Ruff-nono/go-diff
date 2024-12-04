package main

import (
	"sync"
)

var (
	statsGroup *StatsGroup
)

type (
	StatsGroup struct {
		mu         sync.RWMutex
		routeStats map[string]*RouteStats
	}
	RouteStats struct {
		Route             string
		IdenticalCount    int
		DifferenceCount   int
		DifferenceDetails DifferenceDetails
		mu                sync.RWMutex
	}
	DifferenceDetails struct {
		HeaderDifferences map[string]DifferenceExample
		StatusDifferences map[string]DifferenceExample
		BodyDifferences   map[string]DifferenceExample
	}
	DifferenceExample struct {
		Example1 interface{}
		Example2 interface{}
	}
)

func init() {
	statsGroup = &StatsGroup{
		routeStats: make(map[string]*RouteStats),
	}
}

func (sg *StatsGroup) NewRouteStats(routePath string) *RouteStats {
	sg.mu.RLock()
	routeStat, exists := sg.routeStats[routePath]
	sg.mu.RUnlock()
	if exists {
		return routeStat
	}

	sg.mu.Lock()
	defer sg.mu.Unlock()
	if _, exists = sg.routeStats[routePath]; !exists {
		routeStat = &RouteStats{
			Route: routePath,
			DifferenceDetails: DifferenceDetails{
				HeaderDifferences: make(map[string]DifferenceExample),
				StatusDifferences: make(map[string]DifferenceExample),
				BodyDifferences:   make(map[string]DifferenceExample),
			},
		}
		sg.routeStats[routePath] = routeStat
	}
	return routeStat
}

func (sg *StatsGroup) GetAllRouteStats() []*RouteStats {
	routeStatsList := make([]*RouteStats, 0)
	sg.mu.RLock()
	for _, stats := range sg.routeStats {
		routeStatsList = append(routeStatsList, stats)
	}
	sg.mu.RUnlock()
	return routeStatsList
}

func (rs *RouteStats) Add(differences DifferenceDetails) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	hasDifferences := false

	for key, example := range differences.HeaderDifferences {
		hasDifferences = true
		rs.DifferenceDetails.HeaderDifferences[key] = example
	}
	for key, example := range differences.StatusDifferences {
		hasDifferences = true
		rs.DifferenceDetails.StatusDifferences[key] = example
	}
	for key, example := range differences.BodyDifferences {
		hasDifferences = true
		rs.DifferenceDetails.BodyDifferences[key] = example
	}

	if hasDifferences {
		rs.DifferenceCount++
	} else {
		rs.IdenticalCount++
	}
}

func (rs *RouteStats) AddIdentical() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.IdenticalCount++
}

func (rs *RouteStats) GetRouteStats() *RouteStats {
	rs.mu.RLock()
	defer rs.mu.RUnlock()

	headerDifferencesCopy := make(map[string]DifferenceExample, len(rs.DifferenceDetails.HeaderDifferences))
	for k, v := range rs.DifferenceDetails.HeaderDifferences {
		headerDifferencesCopy[k] = v
	}

	statusDifferencesCopy := make(map[string]DifferenceExample, len(rs.DifferenceDetails.StatusDifferences))
	for k, v := range rs.DifferenceDetails.StatusDifferences {
		statusDifferencesCopy[k] = v
	}

	bodyDifferencesCopy := make(map[string]DifferenceExample, len(rs.DifferenceDetails.BodyDifferences))
	for k, v := range rs.DifferenceDetails.BodyDifferences {
		bodyDifferencesCopy[k] = v
	}

	return &RouteStats{
		Route:           rs.Route,
		IdenticalCount:  rs.IdenticalCount,
		DifferenceCount: rs.DifferenceCount,
		DifferenceDetails: DifferenceDetails{
			HeaderDifferences: headerDifferencesCopy,
			StatusDifferences: statusDifferencesCopy,
			BodyDifferences:   bodyDifferencesCopy,
		},
	}
}
