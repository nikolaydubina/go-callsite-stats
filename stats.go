package main

// FuncCallSiteStats is various details about call-site of a function
type FuncCallSiteStats struct {
	ReturnIgnoredCount               uint              `json:"return_ignored_count"`
	ReturnNameCount                  []map[string]uint `json:"return_name_count,omitempty"`
	ArgumentNameCount                []map[string]uint `json:"argument_name_count,omitempty"`
	ArgumentValueCount               []map[string]uint `json:"argument_value_count,omitempty"`
	MultipleAssignmentCount          uint              `json:"multiple_assignment_count"`
	MultipleAssignmentWithOtherCount uint              `json:"multiple_assignment_with_other_count"`
}

func mergeFuncCallSiteStats(from, to *FuncCallSiteStats) {
	to.ReturnNameCount = addSliceCountMap(from.ReturnNameCount, to.ReturnNameCount)
	to.ArgumentNameCount = addSliceCountMap(from.ArgumentNameCount, to.ArgumentNameCount)
	to.ArgumentValueCount = addSliceCountMap(from.ArgumentValueCount, to.ArgumentValueCount)
	to.ReturnIgnoredCount += from.ReturnIgnoredCount
	to.MultipleAssignmentCount += from.MultipleAssignmentCount
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
