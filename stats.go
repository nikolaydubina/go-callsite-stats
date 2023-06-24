package main

type FuncCallSiteStats struct {
	CallCount                        uint              `json:"call_count"`
	ArgumentNameCount                []map[string]uint `json:"argument_name_count,omitempty"`
	ArgumentValueCount               []map[string]uint `json:"argument_value_count,omitempty"`
	ReturnNameCount                  []map[string]uint `json:"return_name_count,omitempty"`
	ReturnIgnoredCount               uint              `json:"return_ignored_count"`
	MultipleAssignmentCount          uint              `json:"multiple_assignment_count"`
	MultipleAssignmentWithOtherCount uint              `json:"multiple_assignment_with_other_count"`
}

func (s *FuncCallSiteStats) IncrBy(from FuncCallSiteStats) {
	s.ReturnNameCount = addSliceCountMap(from.ReturnNameCount, s.ReturnNameCount)
	s.ArgumentNameCount = addSliceCountMap(from.ArgumentNameCount, s.ArgumentNameCount)
	s.ArgumentValueCount = addSliceCountMap(from.ArgumentValueCount, s.ArgumentValueCount)
	s.ReturnIgnoredCount += from.ReturnIgnoredCount
	s.MultipleAssignmentCount += from.MultipleAssignmentCount
	s.CallCount += from.CallCount
}

func addSliceCountMap[T uint](from, to []map[string]T) []map[string]T {
	for i, m := range from {
		if i >= len(to) {
			to = append(to, map[string]T{})
		}
		addCountMap(m, to[i])
	}
	return to
}

func addCountMap[T uint](from, to map[string]T) {
	for k, v := range from {
		to[k] += v
	}
}

type FuncCallSiteStatsMapRepo struct{ m map[FuncID]*FuncCallSiteStats }

func NewFuncCallSiteStatsMapRepo() FuncCallSiteStatsMapRepo {
	return FuncCallSiteStatsMapRepo{m: make(map[FuncID]*FuncCallSiteStats)}
}

func (s FuncCallSiteStatsMapRepo) Add(id FuncID, stats FuncCallSiteStats) {
	if _, ok := s.m[id]; !ok {
		s.m[id] = &FuncCallSiteStats{}
	}
	s.m[id].IncrBy(stats)
}

func (s FuncCallSiteStatsMapRepo) GetAll() map[FuncID]*FuncCallSiteStats { return s.m }
